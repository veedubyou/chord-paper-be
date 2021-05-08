use lapin;
use std::env;

pub fn queue_name() -> String {
    let env_var_result = env::var("RABBITMQ_QUEUE");

    match env_var_result {
        Ok(val) => val,
        Err(err) => panic!("Failed to get queue name from env: {}", err),
    }
}

pub async fn create_connection() -> lapin::Connection {
    let env_var_result = env::var("RABBITMQ_URL");

    let rabbitmq_url = match env_var_result {
        Ok(s) => s,
        Err(err) => panic!("No rabbitMQ url set, error: {}", err),
    };

    let conn_result =
        lapin::Connection::connect(&rabbitmq_url, lapin::ConnectionProperties::default()).await;
    match conn_result {
        Ok(conn) => conn,
        Err(err) => panic!("Failed to connect to rabbitMQ, error: {}", err),
    }
}

pub async fn create_channel(conn: &lapin::Connection) -> lapin::Channel {
    let channel_result = conn.create_channel().await;
    let channel: lapin::Channel = match channel_result {
        Ok(c) => c,
        Err(err) => panic!("Failed to create rabbitMQ channel, error: {}", err),
    };

    let queue_result = channel
        .queue_declare(
            &queue_name(),
            lapin::options::QueueDeclareOptions {
                durable: true,
                passive: false,
                exclusive: false,
                auto_delete: false,
                nowait: false,
            },
            lapin::types::FieldTable::default(),
        )
        .await;

    match queue_result {
        Ok(_) => {}
        Err(err) => panic!("Failed to declare rabbitMQ queue, error: {}", err),
    }

    channel
}
