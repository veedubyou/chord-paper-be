use std::net::SocketAddr;

use crate::server::songs;
use crate::server::users;

use warp::Filter;

pub async fn serve(addr: impl Into<SocketAddr> + 'static) {
    let paths = users::users_server()
        .or(songs::songs_server())
        .with(warp::log("info"));

    warp::serve(paths).run(addr).await;
}
