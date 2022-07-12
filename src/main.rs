mod application;
mod db_utils;
mod environment;
mod rabbitmq_utils;
mod server;

#[tokio::main]
async fn main() {
    pretty_env_logger::init();

    let port = match environment::get_environment() {
        environment::Environment::Development => 5001,
        environment::Environment::Production => 5000,
    };
    
    server::serve(([0, 0, 0, 0], port)).await;
}
