use super::dynamodb;
use super::entity;
use crate::application::concerns::google_verification;
use snafu::Snafu;

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("Failed Google token verification: {}", source))]
    VerificationError { source: google_signin::Error },

    #[snafu(display("Song cannot be created if it already has an ID"))]
    ExistingSongError,

    #[snafu(display("Song owner must be equal to the user persisting the song"))]
    WrongOwnerError,

    #[snafu(display("Data store failed: {}", source))]
    DatastoreError { source: dynamodb::Error },

    #[snafu(display("Song ID not found: {}", id))]
    NotFoundError { id: String },
}

#[derive(Clone)]
pub struct Usecase {
    google_verification: google_verification::GoogleVerification,
    datastore: dynamodb::DynamoDB,
}

impl Usecase {
    pub fn new(
        google_verification: google_verification::GoogleVerification,
        songs_datastore: dynamodb::DynamoDB,
    ) -> Usecase {
        Usecase {
            google_verification: google_verification,
            datastore: songs_datastore,
        }
    }

    pub async fn get_song(&self, id: &str) -> Result<entity::Song, Error> {
        let get_song_result = self.datastore.get_song(id).await;

        match get_song_result {
            Ok(song) => Ok(song),
            Err(err) => Err(map_datastore_error(err, id)),
        }
    }

    pub async fn create_song(
        &self,
        user_id_token: &str,
        mut song: entity::Song,
    ) -> Result<entity::Song, Error> {
        let user = self
            .google_verification
            .verify(user_id_token)
            .map_err(|err| Error::VerificationError { source: err })?;

        if !song.is_new() {
            return Err(Error::ExistingSongError);
        }

        if song.owner != user.id {
            return Err(Error::WrongOwnerError);
        }

        song.create_id();

        match self.datastore.create_song(&song).await {
            Ok(()) => Ok(song),
            Err(err) => Err(map_datastore_error(err, &song.id)),
        }
    }
}

fn map_datastore_error(err: dynamodb::Error, song_id: &str) -> Error {
    match err {
        dynamodb::Error::NotFoundError => Error::NotFoundError {
            id: song_id.to_string(),
        },
        dynamodb::Error::GenericDynamoError { .. }
        | dynamodb::Error::MalformedDataError { .. }
        | dynamodb::Error::SongSerializationError { .. } => Error::DatastoreError { source: err },
    }
}
