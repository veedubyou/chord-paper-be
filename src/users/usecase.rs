use super::entity;
use crate::users::DynamoDB;
use google_signin::IdInfo;
use std::error::Error;

pub struct Usecase {
    client_id: String,
    client: google_signin::Client,
    datastore: DynamoDB,
}

impl Usecase {
    pub fn new(google_client_id: &str, user_datastore: DynamoDB) -> Usecase {
        let mut client = google_signin::Client::new();
        client.audiences.push(google_client_id.to_string());

        Usecase {
            client_id: google_client_id.to_string(),
            client: client,
            datastore: user_datastore,
        }
    }

    pub async fn login(&self, id_token: &str) -> Result<entity::User, Box<dyn Error>> {
        let id_info = self.client.verify(id_token)?;

        let input_user = entity_user_from_google_verification(id_info);
        let output_user = self.datastore.ensure_user(&input_user).await?;

        Ok(output_user)
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
