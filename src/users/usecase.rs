use super::entity;
use crate::users::dynamodb;
use google_signin::IdInfo;
use snafu::Snafu;

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("Failed Google token verification: {}", source))]
    VerificationError { source: google_signin::Error },

    #[snafu(display("Data store failed: {}", source))]
    DatastoreError { source: dynamodb::Error },
}

pub struct Usecase {
    client_id: String,
    client: google_signin::Client,
    datastore: dynamodb::DynamoDB,
}

impl Usecase {
    pub fn new(google_client_id: &str, user_datastore: dynamodb::DynamoDB) -> Usecase {
        let mut client = google_signin::Client::new();
        client.audiences.push(google_client_id.to_string());

        Usecase {
            client_id: google_client_id.to_string(),
            client: client,
            datastore: user_datastore,
        }
    }

    pub async fn login(&self, id_token: &str) -> Result<entity::User, Error> {
        let id_info_result = self.client.verify(id_token);

        let id_info = match id_info_result {
            Ok(id_info) => id_info,
            Err(err) => return Err(Error::VerificationError { source: err }),
        };

        let input_user = entity_user_from_google_verification(id_info);
        let output_user_result = self.datastore.ensure_user(&input_user).await;

        match output_user_result {
            Ok(output_user) => Ok(output_user),
            Err(err) => Err(Error::DatastoreError { source: err }),
        }
    }
}

fn entity_user_from_google_verification(id_info: IdInfo) -> entity::User {
    entity::User {
        id: id_info.sub.to_string(),
        name: id_info.name,
    }
}

impl Clone for Usecase {
    fn clone(&self) -> Usecase {
        Usecase::new(&self.client_id, self.datastore.clone())
    }
}
