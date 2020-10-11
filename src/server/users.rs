use super::cors;
use crate::application::concerns;
use crate::application::users;
use crate::db_utils::dynamodb::db_client;
use warp::Filter;

pub fn users_server() -> warp::filters::BoxedFilter<(impl warp::reply::Reply,)> {
    let users_gateway = create_users_gateway();

    let with_users_gateway = warp::any().map(move || users_gateway.clone());

    warp::post()
        .and(with_users_gateway)
        .and(warp::path!("login"))
        .and(warp::header::<String>("authorization"))
        .and_then(login)
        .with(cors::cors_filter(vec!["POST"]))
        .boxed()
}

fn create_users_gateway() -> users::gateway::Gateway {
    let datastore = users::dynamodb::DynamoDB::new(db_client());
    let verification = concerns::google_verification::GoogleVerification::new();
    let usecase = users::usecase::Usecase::new(verification, datastore);
    users::gateway::Gateway::new(usecase)
}

async fn login(
    users_gateway: users::gateway::Gateway,
    auth_value: String,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(users_gateway.login(&auth_value).await)
}
