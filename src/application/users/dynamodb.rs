use super::entity;
use rusoto_dynamodb::{
    AttributeValue, DynamoDb, UpdateItemError, UpdateItemInput, UpdateItemOutput,
};
use snafu::Snafu;
use std::collections::HashMap;

const TABLE_NAME: &str = "Users";
const ID_FIELD: &str = "id";
const NAME_FIELD: &str = "username";

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("DynamoDB encountered an error: {}", source))]
    DynamoDBUpdateItemError {
        source: rusoto_core::RusotoError<UpdateItemError>,
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

    pub async fn ensure_user(&self, input_user: &entity::User) -> Result<entity::User, Error> {
        let key = {
            let mut keymap: HashMap<String, AttributeValue> = HashMap::new();
            keymap.insert(
                ID_FIELD.to_string(),
                AttributeValue {
                    s: Some(input_user.id.to_string()),
                    b: None,
                    bool: None,
                    bs: None,
                    l: None,
                    m: None,
                    n: None,
                    ns: None,
                    null: None,
                    ss: None,
                },
            );
            keymap
        };

        let user_name_var_name = ":new_name";

        let (expression_attribute, update_expression) = {
            match &input_user.name {
                None => (None, None),
                Some(input_user_name) => {
                    let mut expression_attribute: HashMap<String, AttributeValue> = HashMap::new();
                    expression_attribute.insert(
                        user_name_var_name.to_string(),
                        AttributeValue {
                            s: Some(input_user_name.to_string()),
                            b: None,
                            bool: None,
                            bs: None,
                            l: None,
                            m: None,
                            n: None,
                            ns: None,
                            null: None,
                            ss: None,
                        },
                    );

                    let update_expression = format!("set {} = {}", NAME_FIELD, user_name_var_name);

                    (Some(expression_attribute), Some(update_expression))
                }
            }
        };

        let update_result = self
            .db_client
            .update_item(UpdateItemInput {
                table_name: TABLE_NAME.to_string(),
                return_values: Some("ALL_NEW".to_string()),
                key: key,
                attribute_updates: None,
                condition_expression: None,
                conditional_operator: None,
                expected: None,
                expression_attribute_names: None,
                expression_attribute_values: expression_attribute,
                return_consumed_capacity: None,
                return_item_collection_metrics: None,
                update_expression: update_expression,
            })
            .await;

        match update_result {
            Ok(update_output) => entity_user(update_output),
            Err(err) => Err(Error::DynamoDBUpdateItemError { source: err }),
        }
    }
}

fn entity_user(output: UpdateItemOutput) -> Result<entity::User, Error> {
    let attributes = {
        match output.attributes {
            Some(attributes) => attributes,
            None => {
                return Err(Error::MissingAttributeError {
                    detail: "All attributes are missing".to_string(),
                })
            }
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
