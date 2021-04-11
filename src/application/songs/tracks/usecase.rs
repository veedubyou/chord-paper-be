use super::dynamodb;
use super::entity;
use crate::application::concerns::google_verification;
use snafu::Snafu;

#[derive(Debug, Snafu)]
pub enum Error {
    // #[snafu(display("Failed Google token verification: {}", source))]
    // GoogleVerificationError { source: google_signin::Error },
    //
    // #[snafu(display("Song owner must be equal to the user persisting the tracklist"))]
    // WrongOwnerError,
    #[snafu(display("Data store failed: {}", source))]
    DatastoreError { source: dynamodb::Error },
}

#[derive(Clone)]
pub struct Usecase {
    google_verification: google_verification::GoogleVerification,
    datastore: dynamodb::DynamoDB,
}

impl Usecase {
    pub fn new(
        google_verification: google_verification::GoogleVerification,
        track_datastore: dynamodb::DynamoDB,
    ) -> Usecase {
        Usecase {
            google_verification: google_verification,
            datastore: track_datastore,
        }
    }

    pub async fn get_tracklist(&self, song_id: &str) -> Result<entity::TrackList, Error> {
        if !super::super::entity::Song::is_valid_id(song_id) {
            // if it's not a UUID we won't find it in the datastore
            // just short circuit and don't hit the DB
            return Ok(entity::TrackList::empty(song_id));
        }
        let get_song_result = self.datastore.get_tracklist(song_id).await;

        match get_song_result {
            Ok(song) => Ok(song),
            Err(err) => Err(map_datastore_error(err)),
        }
    }
}

fn map_datastore_error(err: dynamodb::Error) -> Error {
    match err {
        dynamodb::Error::GenericDynamoError { .. } | dynamodb::Error::MalformedDataError { .. } => {
            Error::DatastoreError { source: err }
        }
    }
}
