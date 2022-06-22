mod application;
mod db_utils;
mod environment;
mod server;

#[tokio::main]
async fn main() {
    pretty_env_logger::init();

    server::serve(([0, 0, 0, 0], 5000)).await;
}
