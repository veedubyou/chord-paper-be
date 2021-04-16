use super::dynamodb;
use super::entity;
use crate::application::concerns::user_validation;
use crate::application::songs;
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
}

#[derive(Clone)]
pub struct Usecase {
    user_validation: user_validation::UserValidation,
    songs_datastore: songs::dynamodb::DynamoDB,
    tracks_datastore: dynamodb::DynamoDB,
}

impl Usecase {
    pub fn new(
        user_validation: user_validation::UserValidation,
        track_datastore: dynamodb::DynamoDB,
        songs_datastore: songs::dynamodb::DynamoDB,
    ) -> Usecase {
        Usecase {
            user_validation: user_validation,
            songs_datastore: songs_datastore,
            tracks_datastore: track_datastore,
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
                id: "tracklist.song_id.".to_string(),
            });
        }

        self.verify_song_and_owner(song_id, user_id_token).await?;
        tracklist.ensure_track_ids();

        let set_tracklist_result = self.tracks_datastore.set_tracklist(&tracklist).await;
        match set_tracklist_result {
            Ok(_) => Ok(tracklist),
            Err(err) => Err(Error::DatastoreError { source: err }),
        }
    }
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
