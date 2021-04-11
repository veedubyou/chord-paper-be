use super::entity;
use crate::application::concerns::dynamodb::{deserialize, expression};
use rusoto_dynamodb::{DynamoDb, GetItemInput};
use snafu::Snafu;

const TABLE_NAME: &str = "TrackLists";
const ID_FIELD: &str = "song_id";

#[derive(Debug, Snafu)]
pub enum Error {
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
            return Ok(entity::TrackList::empty(song_id));
        }

        let key = expression::make_single_string_expression(ID_FIELD, song_id);

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
                let result = deserialize::deserialize_from_optional_attributes(output.item);
                map_deserialize_errors(song_id, result)
            }
            Err(rusoto_err) => Err(Error::GenericDynamoError {
                detail: "Failed to get track from data store".to_string(),
                source: Box::new(rusoto_err),
            }),
        }
    }
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
