mod server;
mod usecases;
mod gateway;

#[tokio::main]
async fn main() {
    pretty_env_logger::init();

    server::serve(([127, 0, 0, 1], 5000)).await;
}
