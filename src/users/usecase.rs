use super::entity;
use crate::users::DynamoDB;
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

        let id_str = self.datastore.ensure_user(&id_info.sub).await?;

        Ok(entity::User {
            id: id_str,
            name: id_info.name,
        })
    }
}

impl Clone for Usecase {
    fn clone(&self) -> Usecase {
        Usecase::new(&self.client_id, self.datastore.clone())
    }
}
