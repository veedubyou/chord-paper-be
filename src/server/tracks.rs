use super::cors;
use crate::application::concerns;
use crate::application::songs;
use crate::application::songs::tracks::*;

use crate::db_utils::dynamodb::db_client;
use warp::Filter;

pub fn tracks_server() -> warp::filters::BoxedFilter<(impl warp::reply::Reply,)> {
    let get_tracklist_path = warp::get()
        .and(with_tracks_gateway())
        .and(warp::path!("songs" / String / "tracklist"))
        .and_then(get_tracklist)
        .with(cors::cors_filter(vec!["GET"]))
        .boxed();

    let update_tracklist_path = warp::put()
        .and(with_tracks_gateway())
        .and(warp::header::<String>("authorization"))
        .and(warp::path!("songs" / String / "tracklist"))
        .and(warp::filters::body::json())
        .and_then(put_tracklist)
        .with(cors::cors_filter(vec!["PUT"]))
        .boxed();

    get_tracklist_path.or(update_tracklist_path).boxed()
}

fn with_tracks_gateway() -> warp::filters::BoxedFilter<(gateway::Gateway,)> {
    let datastore = dynamodb::DynamoDB::new(db_client());
    let songs_datastore = songs::dynamodb::DynamoDB::new(db_client());
    let user_validation = concerns::user_validation::UserValidation::new();

    let usecase = usecase::Usecase::new(user_validation, datastore, songs_datastore);
    let gateway = gateway::Gateway::new(usecase);

    warp::any().map(move || gateway.clone()).boxed()
}

async fn get_tracklist(
    tracks_gateway: gateway::Gateway,
    song_id: String,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(tracks_gateway.get_tracklist(&song_id).await)
}

async fn put_tracklist(
    tracks_gateway: gateway::Gateway,
    auth_header_value: String,
    song_id: String,
    tracklist: entity::TrackList,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(tracks_gateway
        .put_tracklist(&auth_header_value, &song_id, tracklist)
        .await)
}
