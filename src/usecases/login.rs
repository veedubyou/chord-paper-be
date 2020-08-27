use serde::Serialize;
use std::error::Error;

#[derive(Serialize)]
pub struct User {
    id: String,
    name: Option<String>,
    email: Option<String>,
}

pub struct Google {
    client: google_signin::Client,
}

impl Google {
    pub fn new(google_client_id: &str) -> Google {
        let mut client = google_signin::Client::new();
        client.audiences.push(google_client_id.to_string());
        Google {
            client: client,
        }
    }

    pub fn verify_login(&self, id_token: &str) -> Result<User, Box<dyn Error>> {
        let id_info = self.client.verify(id_token)?;

        Ok(User{
            id: id_info.sub,
            name: id_info.name,
            email: id_info.email,
        })
    }
}