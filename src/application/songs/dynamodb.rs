use super::entity;
use rusoto_dynamodb::{AttributeValue, DynamoDb, GetItemInput};
use snafu::Snafu;
use std::collections::HashMap;

const TABLE_NAME: &str = "Songs";
const ID_FIELD: &str = "id";

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("Failed to convert entity song to dynamodb format: {}", source))]
    SongSerializationError { source: serde_dynamodb::Error },
    #[snafu(display("Failed to get new song from dynamodb: {}", source))]
    GetItemError {
        source: rusoto_core::RusotoError<rusoto_dynamodb::GetItemError>,
    },
    #[snafu(display("Failed to insert new song to dynamodb format: {}", source))]
    PutItemError {
        source: rusoto_core::RusotoError<rusoto_dynamodb::PutItemError>,
    },
    #[snafu(display("Song ID not found"))]
    NotFoundError,
    #[snafu(display("Data from dynamodb cannot be deserialized into a song: {}", source))]
    MalformedDataError { source: serde_dynamodb::Error },
}

#[derive(Clone)]
pub struct DynamoDB {
    db_client: rusoto_dynamodb::DynamoDbClient,
}

impl DynamoDB {
    pub fn new(db_client: rusoto_dynamodb::DynamoDbClient) -> DynamoDB {
        DynamoDB {
            db_client: db_client,
        }
    }

    pub async fn get_song(&self, id: &str) -> Result<entity::Song, Error> {
        let key = {
            let mut map: HashMap<String, AttributeValue> = HashMap::new();
            map.insert(
                ID_FIELD.to_string(),
                AttributeValue {
                    s: Some(id.to_string()),
                    ..Default::default()
                },
            );
            map
        };

        let get_result = self
            .db_client
            .get_item(GetItemInput {
                key: key,
                table_name: TABLE_NAME.to_string(),
                consistent_read: Some(true),
                ..Default::default()
            })
            .await;

        println!("{:#?}", get_result);

        match get_result {
            Ok(output) => song_from_attributes(output.item),
            Err(rusoto_err) => {
                if let rusoto_core::RusotoError::Service(err) = &rusoto_err {
                    if let rusoto_dynamodb::GetItemError::ResourceNotFound(_) = err {
                        return Err(Error::NotFoundError);
                    }
                }

                Err(Error::GetItemError { source: rusoto_err })
            }
        }
    }

    pub async fn create_song(&self, song: &entity::Song) -> Result<(), Error> {
        let dynamodb_item = serde_dynamodb::to_hashmap(&song)
            .map_err(|err| Error::SongSerializationError { source: err })?;

        let put_result = self
            .db_client
            .put_item(rusoto_dynamodb::PutItemInput {
                table_name: TABLE_NAME.to_string(),
                item: dynamodb_item,
                condition_expression: Some("attribute_not_exists(id)".to_string()),
                ..Default::default()
            })
            .await;

        match put_result {
            Ok(_) => Ok(()),
            Err(err) => Err(Error::PutItemError { source: err }),
        }
    }
}

fn song_from_attributes(
    output: Option<HashMap<String, AttributeValue>>,
) -> Result<entity::Song, Error> {
    let map = output.ok_or(Error::NotFoundError)?;
    let song: entity::Song = serde_dynamodb::from_hashmap(map)
        .map_err(|err| Error::MalformedDataError { source: err })?;
    Ok(song)
}
