use super::super::songs;
use super::super::users;

use google_signin::IdInfo;
use snafu::Snafu;

const GOOGLE_CLIENT_ID: &str =
    "650853277550-ta69qbfcvdl6tb5ogtnh2d07ae9rcdlf.apps.googleusercontent.com";

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("Failed Google token verification: {}", source))]
    GoogleVerificationError { source: google_signin::Error },
    #[snafu(display("This user does not own the requested song"))]
    WrongOwnerError,
}

pub struct UserValidation {
    client: google_signin::Client,
}

impl UserValidation {
    pub fn new() -> UserValidation {
        let mut client = google_signin::Client::new();
        client.audiences.push(GOOGLE_CLIENT_ID.to_string());
        UserValidation { client: client }
    }

    pub fn verify_user(&self, id_token: &str) -> Result<users::entity::User, Error> {
        match self.client.verify(id_token) {
            Ok(id_info) => Ok(entity_user_from_google_verification(id_info)),
            Err(err) => Err(Error::GoogleVerificationError { source: err }),
        }
    }

    pub fn verify_owner(
        &self,
        id_token: &str,
        song_summary: &songs::entity::SongSummary,
    ) -> Result<(), Error> {
        let user = self.verify_user(id_token)?;

        if song_summary.owner != user.id {
            return Err(Error::WrongOwnerError);
        }

        Ok(())
    }
}

fn entity_user_from_google_verification(id_info: IdInfo) -> users::entity::User {
    users::entity::User {
        id: id_info.sub.to_string(),
        name: id_info.name,
    }
}

impl Clone for UserValidation {
    fn clone(&self) -> UserValidation {
        UserValidation::new()
    }
}
