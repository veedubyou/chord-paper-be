use crate::users;

const GOOGLE_CLIENT_ID: &str =
    "650853277550-ta69qbfcvdl6tb5ogtnh2d07ae9rcdlf.apps.googleusercontent.com";

pub fn create_users_gateway() -> users::Gateway {
    let usecase = users::Usecase::new(GOOGLE_CLIENT_ID);
    users::Gateway::new(usecase)
}
