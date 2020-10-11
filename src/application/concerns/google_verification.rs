use super::super::users::entity;
use google_signin::IdInfo;

const GOOGLE_CLIENT_ID: &str =
    "650853277550-ta69qbfcvdl6tb5ogtnh2d07ae9rcdlf.apps.googleusercontent.com";

pub struct GoogleVerification {
    client: google_signin::Client,
}

impl GoogleVerification {
    pub fn new() -> GoogleVerification {
        let mut client = google_signin::Client::new();
        client.audiences.push(GOOGLE_CLIENT_ID.to_string());
        GoogleVerification { client: client }
    }

    pub fn verify(&self, id_token: &str) -> Result<entity::User, google_signin::Error> {
        let id_info = self.client.verify(id_token)?;

        Ok(entity_user_from_google_verification(id_info))
    }
}

fn entity_user_from_google_verification(id_info: IdInfo) -> entity::User {
    entity::User {
        id: id_info.sub.to_string(),
        name: id_info.name,
    }
}

impl Clone for GoogleVerification {
    fn clone(&self) -> GoogleVerification {
        GoogleVerification::new()
    }
}
