use super::dynamodb;
use super::entity;
use crate::application::concerns::user_validation;
use crate::application::songs;

use snafu::Snafu;

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("Failed user validation: {}", source))]
    UserValidationError { source: user_validation::Error },

    #[snafu(display("This authenticated user cannot access this user's resources"))]
    OwnerVerificationError,

    #[snafu(display("Data store failed: {}", source))]
    DatastoreError { source: Box<dyn std::error::Error> },
}

#[derive(Clone)]
pub struct Usecase {
    google_verification: user_validation::UserValidation,
    users_datastore: dynamodb::DynamoDB,
    songs_datastore: songs::dynamodb::DynamoDB,
}

impl Usecase {
    pub fn new(
        google_verification: user_validation::UserValidation,
        users_datastore: dynamodb::DynamoDB,
        songs_datastore: songs::dynamodb::DynamoDB,
    ) -> Usecase {
        Usecase {
            google_verification,
            users_datastore,
            songs_datastore,
        }
    }

    pub async fn login(&self, google_id_token: &str) -> Result<entity::User, Error> {
        let input_user = self.verify_user(google_id_token)?;

        let output_user_result = self.users_datastore.ensure_user(&input_user).await;

        match output_user_result {
            Ok(output_user) => Ok(output_user),
            Err(err) => Err(Error::DatastoreError {
                source: Box::new(err),
            }),
        }
    }

    pub async fn songs_for_user(
        &self,
        google_id_token: &str,
        owner_id: &str,
    ) -> Result<Vec<songs::entity::SongSummary>, Error> {
        let user = self.verify_user(google_id_token)?;

        if user.id != owner_id {
            return Err(Error::OwnerVerificationError);
        }

        let songs_result = self
            .songs_datastore
            .get_song_summaries_for_owner_id(&user.id)
            .await;

        match songs_result {
            Ok(songs_list) => Ok(songs_list),
            Err(err) => Err(Error::DatastoreError {
                source: Box::new(err),
            }),
        }
    }

    fn verify_user(&self, google_id_token: &str) -> Result<entity::User, Error> {
        self.google_verification
            .verify_user(google_id_token)
            .map_err(|err| Error::UserValidationError { source: err })
    }
}
