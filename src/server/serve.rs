use std::net::SocketAddr;
use warp::Filter;
use crate::gateway;

const GOOGLE_CLIENT_ID: &str = "650853277550-ta69qbfcvdl6tb5ogtnh2d07ae9rcdlf.apps.googleusercontent.com";

pub async fn serve(addr: impl Into<SocketAddr> + 'static) {
    let verify_login = warp::post()
        .and(warp::path("login"))
        .and(warp::body::content_length_limit(1024 * 16))
        .and(warp::body::json())
        .map(|request: crate::gateway::login::VerifyLoginRequest| {
            let gateway = gateway::login::Google::new(GOOGLE_CLIENT_ID);
            gateway.verify_login(request)
        });

    let paths = verify_login;

    warp::serve(paths)
        .run(addr)
        .await;
}
