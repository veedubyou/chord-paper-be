use crate::db_utils::dynamodb::db_client;
use crate::users;

const GOOGLE_CLIENT_ID: &str =
    "650853277550-ta69qbfcvdl6tb5ogtnh2d07ae9rcdlf.apps.googleusercontent.com";

pub fn create_users_gateway() -> users::Gateway {
    let datastore = users::DynamoDB::new(db_client());
    let usecase = users::Usecase::new(GOOGLE_CLIENT_ID, datastore);
    users::Gateway::new(usecase)
}

pub async fn login(
    users_gateway: users::Gateway,
    user_id: String,
    auth_value: String,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(users_gateway.login(&user_id, &auth_value).await)
}
