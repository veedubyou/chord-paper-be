use super::cors;
use crate::application::concerns;
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

    get_tracklist_path.boxed()
}

fn with_tracks_gateway() -> warp::filters::BoxedFilter<(gateway::Gateway,)> {
    let datastore = dynamodb::DynamoDB::new(db_client());
    let verification = concerns::google_verification::GoogleVerification::new();
    let usecase = usecase::Usecase::new(verification, datastore);
    let gateway = gateway::Gateway::new(usecase);

    warp::any().map(move || gateway.clone()).boxed()
}

async fn get_tracklist(
    songs_gateway: gateway::Gateway,
    song_id: String,
) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    Ok(songs_gateway.get_tracklist(&song_id).await)
}
