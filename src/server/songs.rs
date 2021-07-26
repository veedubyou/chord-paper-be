use super::cors;
use crate::application::concerns;
use crate::application::songs::*;
use crate::db_utils::dynamodb::db_client;
use warp::Filter;

pub fn songs_server() -> warp::filters::BoxedFilter<(impl warp::reply::Reply,)> {
    let get_song_path = warp::get()
        .and(with_songs_gateway())
        .and(warp::path!("songs" / String))
        .and_then(get_song)
        .with(cors::cors_filter(vec!["GET"]))
        .boxed();

    let create_song_path = warp::post()
        .and(with_songs_gateway())
        .and(warp::header::<String>("authorization"))
        .and(warp::path!("songs"))
        .and(warp::filters::body::json())
        .and_then(create_song)
        .with(cors::cors_filter(vec!["POST"]))
        .boxed();

    let update_song_path = warp::put()
        .and(with_songs_gateway())
        .and(warp::header::<String>("authorization"))
        .and(warp::path!("songs" / String))
        .and(warp::filters::body::json())
        .and_then(update_song)
        .with(cors::cors_filter(vec!["PUT"]))
        .boxed();

    let delete_song_path = warp::delete()
        .and(with_songs_gateway())
        .and(warp::header::<String>("authorization"))
        .and(warp::path!("songs" / String))
        .and_then(delete_song)
        .with(cors::cors_filter(vec!["DELETE"]))
        .boxed();

    get_song_path
        .or(create_song_path)
        .or(update_song_path)
        .or(delete_song_path)
        .boxed()
}

fn with_songs_gateway() -> warp::filters::BoxedFilter<(gateway::Gateway,)> {
    let datastore = dynamodb::DynamoDB::new(db_client());
    let verification = concerns::user_validation::UserValidation::new();
    let usecase = usecase::Usecase::new(verification, datastore);
    let gateway = gateway::Gateway::new(usecase);

    warp::any().map(move || gateway.clone()).boxed()
}

async fn get_song(
    songs_gateway: gateway::Gateway,
    song_id: String,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(songs_gateway.get_song(&song_id).await)
}

async fn create_song(
    songs_gateway: gateway::Gateway,
    auth_header_value: String,
    song: entity::Song,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(songs_gateway.create_song(&auth_header_value, song).await)
}

async fn update_song(
    songs_gateway: gateway::Gateway,
    auth_header_value: String,
    song_id: String,
    song: entity::Song,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(songs_gateway
        .update_song(&auth_header_value, &song_id, song)
        .await)
}

async fn delete_song(
    songs_gateway: gateway::Gateway,
    auth_header_value: String,
    song_id: String,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(songs_gateway
        .delete_song(&auth_header_value, &song_id)
        .await)
}
