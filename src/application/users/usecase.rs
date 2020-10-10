use super::super::concerns::google_verification;
use super::dynamodb;
use super::entity;
use snafu::Snafu;

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("Failed Google token verification: {}", source))]
    VerificationError { source: google_signin::Error },

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
        user_datastore: dynamodb::DynamoDB,
    ) -> Usecase {
        Usecase {
            google_verification,
            datastore: user_datastore,
        }
    }

    pub async fn login(&self, user_id_token: &str) -> Result<entity::User, Error> {
        let input_user = self
            .google_verification
            .verify(user_id_token)
            .map_err(|err| Error::VerificationError { source: err })?;

        let output_user_result = self.datastore.ensure_user(&input_user).await;

        match output_user_result {
            Ok(output_user) => Ok(output_user),
            Err(err) => Err(Error::DatastoreError { source: err }),
        }
    }
}
