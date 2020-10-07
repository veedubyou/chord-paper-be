use super::entity::User;
use std::error::Error;

pub struct Usecase {
    client_id: String,
    client: google_signin::Client,
}

impl Usecase {
    pub fn new(google_client_id: &str) -> Usecase {
        let mut client = google_signin::Client::new();
        client.audiences.push(google_client_id.to_string());

        Usecase {
            client_id: google_client_id.to_string(),
            client: client,
        }
    }

    pub fn login(&self, id_token: &str) -> Result<User, Box<dyn Error>> {
        let id_info = self.client.verify(id_token)?;

        Ok(User {
            id: id_info.sub,
            name: id_info.name,
        })
    }
}

impl Clone for Usecase {
    fn clone(&self) -> Usecase {
        Usecase::new(&self.client_id)
    }
}
