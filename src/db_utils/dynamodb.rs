use crate::environment;
use rusoto_core;
use rusoto_dynamodb;

// for now, until DB is implemented
#[allow(dead_code)]
pub fn db_client() -> rusoto_dynamodb::DynamoDbClient {
    match environment::get_environment() {
        environment::Environment::Production => aws_client(),
        environment::Environment::Development => local_client(),
    }
}

fn aws_client() -> rusoto_dynamodb::DynamoDbClient {
    rusoto_dynamodb::DynamoDbClient::new(rusoto_core::Region::UsWest1)
}

fn local_client() -> rusoto_dynamodb::DynamoDbClient {
    let region = rusoto_core::Region::Custom {
        name: "local-test".to_owned(),
        endpoint: "http://localhost:8000".to_owned(),
    };

    rusoto_dynamodb::DynamoDbClient::new(region)
}
