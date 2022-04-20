use super::dynamodb;
use super::entity;
use crate::application::concerns::user_validation;
use crate::application::songs;

use snafu::Snafu;

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("User account does not exist: {}", source))]
    NoAccountError { source: Box<dyn std::error::Error> },

    #[snafu(display("Failed Google validation: {}, {}", detail, source))]
    GoogleValidationError {
        detail: String,
        source: Box<dyn std::error::Error>,
    },

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

        let output_user_result = self.users_datastore.get_user(&input_user.id).await;

        match output_user_result {
            Ok(output_user) => Ok(output_user),

            Err(datastore_err) => {
                if let dynamodb::Error::NotFoundError = datastore_err {
                    return Err(Error::NoAccountError {
                        source: Box::new(datastore_err),
                    });
                }

                Err(Error::DatastoreError {
                    source: Box::new(datastore_err),
                })
            }
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
            .map_err(|err| Error::GoogleValidationError {
                detail: "Failed Google user verification".to_string(),
                source: Box::new(err),
            })
    }
}
