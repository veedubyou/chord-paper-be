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

    #[snafu(display("{}: {}", detail, source))]
    GenericDynamoError {
        detail: String,
        source: Box<dyn std::error::Error>,
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
        if id.is_empty() {
            return Err(Error::NotFoundError);
        }

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

        match get_result {
            Ok(output) => song_from_attributes(output.item),
            Err(rusoto_err) => {
                if let rusoto_core::RusotoError::Service(err) = &rusoto_err {
                    if let rusoto_dynamodb::GetItemError::ResourceNotFound(_) = err {
                        return Err(Error::NotFoundError);
                    }
                }

                Err(Error::GenericDynamoError {
                    detail: "Failed to get song from data store".to_string(),
                    source: Box::new(rusoto_err),
                })
            }
        }
    }

    pub async fn create_song(&self, song: &entity::Song) -> Result<(), Error> {
        let dynamodb_item = dynamodb_item_from_song(song)?;

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
            Err(err) => Err(Error::GenericDynamoError {
                detail: "Failed to put song into data store".to_string(),
                source: Box::new(err),
            }),
        }
    }

    pub async fn update_song(&self, song: &entity::Song) -> Result<(), Error> {
        if song.id.is_empty() {
            return Err(Error::NotFoundError);
        }

        let dynamodb_item = dynamodb_item_from_song(song)?;
        let put_result = self
            .db_client
            .put_item(rusoto_dynamodb::PutItemInput {
                table_name: TABLE_NAME.to_string(),
                item: dynamodb_item,
                condition_expression: Some("attribute_exists(id)".to_string()),
                ..Default::default()
            })
            .await;

        match put_result {
            Ok(_) => Ok(()),
            Err(err) => Err(Error::GenericDynamoError {
                detail: "Failed to update song in data store".to_string(),
                source: Box::new(err),
            }),
        }
    }
}

fn dynamodb_item_from_song(song: &entity::Song) -> Result<HashMap<String, AttributeValue>, Error> {
    serde_dynamodb::to_hashmap(&song).map_err(|err| Error::SongSerializationError { source: err })
}

fn song_from_attributes(
    output: Option<HashMap<String, AttributeValue>>,
) -> Result<entity::Song, Error> {
    let map = output.ok_or(Error::NotFoundError)?;
    let song: entity::Song = serde_dynamodb::from_hashmap(map)
        .map_err(|err| Error::MalformedDataError { source: err })?;
    Ok(song)
}
