use super::dynamodb;
use super::entity;
use crate::application::concerns::user_validation;
use crate::application::songs;
use crate::application::songs::tracks::entity::TrackList;
use amq_protocol_types::ShortString;
use lapin;
use serde::{Deserialize, Serialize};
use snafu::Snafu;

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("An invalid ID is provided, id: {}", id))]
    InvalidIDError { id: String },
    #[snafu(display("Failed Google token verification: {}", source))]
    GoogleVerificationError { source: google_signin::Error },
    #[snafu(display("Song owner must be equal to the user persisting the tracklist"))]
    WrongOwnerError,
    #[snafu(display("Data store failed: {}", source))]
    DatastoreError { source: dynamodb::Error },
    #[snafu(display("Failed to fetch song for this track: {}", source))]
    GetSongError { source: songs::dynamodb::Error },
    #[snafu(display("Failed to publish worker message: {}", msg))]
    PublishError { msg: String },
}

#[derive(Clone)]
pub struct Usecase {
    user_validation: user_validation::UserValidation,
    songs_datastore: songs::dynamodb::DynamoDB,
    tracks_datastore: dynamodb::DynamoDB,
    rabbitmq_channel: lapin::Channel,
    queue_name: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct SplitRequestTrack {
    pub tracklist_id: String,
    pub track_id: String,
}

impl Usecase {
    pub fn new(
        user_validation: user_validation::UserValidation,
        tracks_datastore: dynamodb::DynamoDB,
        songs_datastore: songs::dynamodb::DynamoDB,
        rabbitmq_channel: lapin::Channel,
        queue_name: &str,
    ) -> Usecase {
        Usecase {
            user_validation: user_validation,
            songs_datastore: songs_datastore,
            tracks_datastore: tracks_datastore,
            rabbitmq_channel: rabbitmq_channel,
            queue_name: queue_name.to_string(),
        }
    }

    pub async fn get_tracklist(&self, song_id: &str) -> Result<entity::TrackList, Error> {
        if !super::super::entity::Song::is_valid_id(song_id) {
            // if it's not a UUID we won't find it in the datastore
            // just short circuit and don't hit the DB
            return Ok(entity::TrackList::empty(song_id));
        }

        let get_tracklist_result = self.tracks_datastore.get_tracklist(song_id).await;
        get_tracklist_result.map_err(|err| map_datastore_error(err))
    }

    async fn verify_song_and_owner(&self, song_id: &str, user_id_token: &str) -> Result<(), Error> {
        let song_result = self.songs_datastore.get_song(song_id).await;
        let song = song_result.map_err(|err| Error::GetSongError { source: err })?;

        if song_id != song.summary.id {
            return Err(Error::InvalidIDError {
                id: song_id.to_string(),
            });
        }

        let validation_result = self
            .user_validation
            .verify_owner(user_id_token, &song.summary);

        validation_result.map_err(map_user_validation_error)
    }

    pub async fn set_tracklist(
        &self,
        user_id_token: &str,
        song_id: &str,
        mut tracklist: entity::TrackList,
    ) -> Result<entity::TrackList, Error> {
        if !tracklist.has_valid_id() {
            return Err(Error::InvalidIDError {
                id: tracklist.song_id.to_string(),
            });
        }

        self.verify_song_and_owner(song_id, user_id_token).await?;
        let split_requests = ensure_track_ids_and_collect_split_requests(&mut tracklist);

        // crazy futures nonsense - need to keep awaits separated
        {
            let set_tracklist_result = self.tracks_datastore.set_tracklist(&tracklist).await;
            set_tracklist_result.map_err(|err| Error::DatastoreError { source: err })?;
        }
        {
            for split_request in split_requests.iter() {
                self.publish_split_job(split_request).await?;
            }
        }

        Ok(tracklist)
    }

    async fn publish_split_job(&self, split_request: &SplitRequestTrack) -> Result<(), Error> {
        let serialize_result = serde_json::to_vec(split_request);

        let payload = serialize_result.map_err(|_| Error::PublishError {
            msg: "Failed to serialize message payload".to_string(),
        })?;

        let mut publish_properties = lapin::BasicProperties::default();
        publish_properties = publish_properties.with_kind(ShortString::from("transfer_original"));
        publish_properties =
            publish_properties.with_content_encoding(ShortString::from("application/json"));

        let publish_result = self
            .rabbitmq_channel
            .basic_publish(
                "",
                &self.queue_name,
                lapin::options::BasicPublishOptions {
                    mandatory: true,
                    immediate: false,
                },
                payload,
                publish_properties,
            )
            .await;

        match publish_result {
            Ok(_) => Ok(()),
            Err(_) => Err(Error::PublishError {
                msg: "Failed to publish message to RabbitMQ".to_string(),
            }),
        }
    }
}

fn ensure_track_ids_and_collect_split_requests(
    tracklist: &mut TrackList,
) -> Vec<SplitRequestTrack> {
    let mut split_requests: Vec<SplitRequestTrack> = vec![];
    for track in &mut tracklist.tracks {
        if track.is_new() {
            track.create_id();

            if track.is_split_request() {
                split_requests.push(SplitRequestTrack {
                    tracklist_id: tracklist.song_id.to_string(),
                    track_id: track.id.to_string(),
                })
            }
        }
    }

    split_requests
}

fn map_datastore_error(err: dynamodb::Error) -> Error {
    match err {
        dynamodb::Error::GenericDynamoError { .. }
        | dynamodb::Error::MalformedDataError { .. }
        | dynamodb::Error::TrackListSerializationError { .. } => {
            Error::DatastoreError { source: err }
        }
        dynamodb::Error::InvalidIDError { id } => Error::InvalidIDError { id },
    }
}

fn map_user_validation_error(err: user_validation::Error) -> Error {
    match err {
        user_validation::Error::GoogleVerificationError { source } => {
            Error::GoogleVerificationError { source }
        }
        user_validation::Error::WrongOwnerError => Error::WrongOwnerError,
    }
}
