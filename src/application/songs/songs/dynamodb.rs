use super::entity;
use crate::application::concerns::dynamodb::{deserialize, expression};
use rusoto_dynamodb::{AttributeValue, DynamoDb, GetItemInput, QueryInput};
use serde::Deserialize;
use snafu::Snafu;
use std::collections::HashMap;

const TABLE_NAME: &str = "Songs";
const OWNER_INDEX: &str = "owner-index";
const ID_FIELD: &str = "id";
const OWNER_FIELD: &str = "owner";

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

        let key = expression::make_single_string(ID_FIELD, id);

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
            Ok(output) => {
                let result = deserialize::from_optional_attributes(output.item);
                map_deserialize_errors(result)
            }
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

    pub async fn get_song_summaries_for_owner_id(
        &self,
        owner_id: &str,
    ) -> Result<Vec<entity::SongSummary>, Error> {
        if owner_id.is_empty() {
            return Err(Error::NotFoundError);
        }

        const OWNER_FIELD_VAR_NAME: &str = "#owner";
        const OWNER_FIELD_VAR_VALUE: &str = ":owner";

        let owner_expression = expression::make_single_string(OWNER_FIELD_VAR_VALUE, owner_id);
        let owner_name =
            expression::make_hashmap(OWNER_FIELD_VAR_NAME.to_string(), OWNER_FIELD.to_string());

        let query_result = self
            .db_client
            .query(QueryInput {
                table_name: TABLE_NAME.to_string(),
                index_name: Some(OWNER_INDEX.to_string()),
                key_condition_expression: Some(format!(
                    "{} = {}",
                    OWNER_FIELD_VAR_NAME, OWNER_FIELD_VAR_VALUE
                )),
                expression_attribute_names: Some(owner_name),
                expression_attribute_values: Some(owner_expression),
                ..Default::default()
            })
            .await;

        match query_result {
            Ok(output) => {
                let result = deserialize::from_optional_list_attributes(output.items);
                map_deserialize_errors(result)
            }
            Err(err) => Err(Error::GenericDynamoError {
                detail: "Failed to query all songs from data store".to_string(),
                source: Box::new(err),
            }),
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
        if song.is_new() {
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

    pub async fn delete_song(&self, song_id: &str) -> Result<(), Error> {
        let key = expression::make_single_string("id", song_id);

        let delete_result = self
            .db_client
            .delete_item(rusoto_dynamodb::DeleteItemInput {
                key: key,
                table_name: TABLE_NAME.to_string(),
                ..Default::default()
            })
            .await;

        match delete_result {
            Ok(_) => Ok(()),
            Err(err) => Err(Error::GenericDynamoError {
                detail: "Failed to delete song in data store".to_string(),
                source: Box::new(err),
            }),
        }
    }
}

fn dynamodb_item_from_song(song: &entity::Song) -> Result<HashMap<String, AttributeValue>, Error> {
    serde_dynamodb::to_hashmap(&song).map_err(|err| Error::SongSerializationError { source: err })
}

fn map_deserialize_errors<'a, T: Deserialize<'a>>(
    result: Result<T, deserialize::Error>,
) -> Result<T, Error> {
    result.map_err(|err| match err {
        deserialize::Error::NotFoundError => Error::NotFoundError,
        deserialize::Error::MalformedDataError { source } => {
            Error::MalformedDataError { source: source }
        }
    })
}
