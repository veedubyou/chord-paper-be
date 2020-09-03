mod server;
mod usecases;
mod gateway;

#[tokio::main]
async fn main() {
    pretty_env_logger::init();

    server::serve(([0, 0, 0, 0], 5000)).await;
}
