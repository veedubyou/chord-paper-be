use super::cors;
use crate::application::concerns;
use crate::application::songs;
use crate::application::users;

use crate::db_utils::dynamodb::db_client;
use warp::Filter;

pub fn users_server() -> warp::filters::BoxedFilter<(impl warp::reply::Reply,)> {
    let login_path = warp::post()
        .and(with_users_gateway())
        .and(warp::header::<String>("authorization"))
        .and(warp::path!("login"))
        .and_then(login)
        .with(cors::cors_filter(vec!["POST"]));

    let songs_for_user = warp::get()
        .and(with_users_gateway())
        .and(warp::header::<String>("authorization"))
        .and(warp::path!("users" / String / "songs"))
        .and_then(songs_for_user)
        .with(cors::cors_filter(vec!["GET"]));

    login_path.or(songs_for_user).boxed()
}

fn with_users_gateway() -> warp::filters::BoxedFilter<(users::gateway::Gateway,)> {
    let users_datastore = users::dynamodb::DynamoDB::new(db_client());
    let songs_datastore = songs::dynamodb::DynamoDB::new(db_client());
    let verification = concerns::user_validation::UserValidation::new();
    let usecase = users::usecase::Usecase::new(verification, users_datastore, songs_datastore);
    let gateway = users::gateway::Gateway::new(usecase);

    warp::any().map(move || gateway.clone()).boxed()
}

async fn login(
    users_gateway: users::gateway::Gateway,
    auth_value: String,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(users_gateway.login(&auth_value).await)
}

async fn songs_for_user(
    users_gateway: users::gateway::Gateway,
    auth_value: String,
    user_id: String,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(users_gateway.songs_for_user(&auth_value, &user_id).await)
}
