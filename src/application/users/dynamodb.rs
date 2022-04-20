use super::entity;
use crate::application::concerns::dynamodb::expression;
use rusoto_dynamodb::{AttributeValue, DynamoDb, GetItemInput, GetItemOutput};
use snafu::Snafu;
use std::collections::HashMap;

const TABLE_NAME: &str = "Users";
const ID_FIELD: &str = "id";
const NAME_FIELD: &str = "username";

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("User ID not found"))]
    NotFoundError,

    #[snafu(display("{}: {}", detail, source))]
    GenericDynamoError {
        detail: String,
        source: Box<dyn std::error::Error>,
    },

    #[snafu(display("An attribute in the data record is missing: {}", detail))]
    MissingAttributeError { detail: String },

    #[snafu(display(
        "An attribute in the data record does not have the expected type: {}",
        detail
    ))]
    WrongAttributeTypeError { detail: String },
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

    pub async fn get_user(&self, user_id: &str) -> Result<entity::User, Error> {
        if user_id.is_empty() {
            return Err(Error::NotFoundError);
        }

        let key = expression::make_single_string(ID_FIELD, user_id);

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
            Ok(output) => entity_user(output),

            Err(rusoto_err) => {
                if let rusoto_core::RusotoError::Service(err) = &rusoto_err {
                    if let rusoto_dynamodb::GetItemError::ResourceNotFound(_) = err {
                        return Err(Error::NotFoundError);
                    }
                }

                Err(Error::GenericDynamoError {
                    detail: "Failed to get user from data store".to_string(),
                    source: Box::new(rusoto_err),
                })
            }
        }
    }
}

fn entity_user(output: GetItemOutput) -> Result<entity::User, Error> {
    let attributes = {
        match output.item {
            Some(attributes) => attributes,
            None => return Err(Error::NotFoundError),
        }
    };

    Ok(entity::User {
        id: entity_user_id(&attributes)?,
        name: entity_user_name(&attributes)?,
    })
}

fn entity_user_id(attributes: &HashMap<String, AttributeValue>) -> Result<String, Error> {
    let id_value = {
        match attributes.get(ID_FIELD) {
            Some(attribute_value) => attribute_value,
            None => {
                return Err(Error::MissingAttributeError {
                    detail: "No id found in user".to_string(),
                })
            }
        }
    };

    match &id_value.s {
        Some(id_str) => Ok(id_str.to_string()),
        None => {
            return Err(Error::WrongAttributeTypeError {
                detail: "ID attribute on user is not a string".to_string(),
            })
        }
    }
}

fn entity_user_name(attributes: &HashMap<String, AttributeValue>) -> Result<Option<String>, Error> {
    let name_value = {
        match attributes.get(NAME_FIELD) {
            Some(attribute_value) => attribute_value,
            None => return Ok(None),
        }
    };

    match &name_value.s {
        Some(name_str) => Ok(Some(name_str.to_string())),
        None => {
            return Err(Error::WrongAttributeTypeError {
                detail: "Name attribute on user is not a string".to_string(),
            })
        }
    }
}
