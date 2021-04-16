use super::entity;
use crate::application::concerns::dynamodb::{deserialize, expression};
use rusoto_dynamodb::{AttributeValue, DynamoDb, GetItemInput};
use snafu::Snafu;
use std::collections::HashMap;

const TABLE_NAME: &str = "TrackLists";
const ID_FIELD: &str = "song_id";

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("Failed to convert entity tracklist to dynamodb format: {}", source))]
    TrackListSerializationError { source: serde_dynamodb::Error },

    #[snafu(display("An invalid ID is provided, id: {}", id))]
    InvalidIDError { id: String },

    #[snafu(display("{}: {}", detail, source))]
    GenericDynamoError {
        detail: String,
        source: Box<dyn std::error::Error>,
    },

    #[snafu(display(
        "Data from dynamodb cannot be deserialized into a TrackList: {}",
        source
    ))]
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

    pub async fn get_tracklist(&self, song_id: &str) -> Result<entity::TrackList, Error> {
        if song_id.is_empty() {
            return Err(Error::InvalidIDError {
                id: song_id.to_string(),
            });
        }

        let key = expression::make_single_string(ID_FIELD, song_id);

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
                println!("{:#?}", output.item);

                let result = deserialize::from_optional_attributes(output.item);
                map_deserialize_errors(song_id, result)
            }
            Err(rusoto_err) => Err(Error::GenericDynamoError {
                detail: "Failed to get track from data store".to_string(),
                source: Box::new(rusoto_err),
            }),
        }
    }

    pub async fn set_tracklist(&self, tracklist: &entity::TrackList) -> Result<(), Error> {
        if !tracklist.has_valid_id() {
            return Err(Error::InvalidIDError { id: "".to_string() });
        }

        let dynamodb_item = serialize_tracklist(tracklist)?;

        let put_result = self
            .db_client
            .put_item(rusoto_dynamodb::PutItemInput {
                table_name: TABLE_NAME.to_string(),
                item: dynamodb_item,
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

fn serialize_tracklist(
    tracklist: &entity::TrackList,
) -> Result<HashMap<String, AttributeValue>, Error> {
    serde_dynamodb::to_hashmap(&tracklist)
        .map_err(|err| Error::TrackListSerializationError { source: err })
}

fn map_deserialize_errors(
    song_id: &str,
    result: Result<entity::TrackList, deserialize::Error>,
) -> Result<entity::TrackList, Error> {
    match result {
        Ok(track_list) => Ok(track_list),
        Err(err) => match err {
            deserialize::Error::NotFoundError => Ok(entity::TrackList::empty(song_id)),
            deserialize::Error::MalformedDataError { source } => {
                Err(Error::MalformedDataError { source: source })
            }
        },
    }
}
