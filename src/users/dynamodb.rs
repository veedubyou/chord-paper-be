use rusoto_dynamodb::{AttributeValue, DynamoDb, UpdateItemInput, UpdateItemOutput};
use std::collections::HashMap;
use std::error::Error;
use std::fmt;

const ID_FIELD: &str = "id";

#[derive(Debug, Clone)]
struct DBError {
    detail: String,
}

impl fmt::Display for DBError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.detail)
    }
}

impl Error for DBError {}

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

    pub async fn ensure_user(&self, input_id: &str) -> Result<String, Box<dyn Error>> {
        let key = {
            let mut keymap: HashMap<String, AttributeValue> = HashMap::new();
            keymap.insert(
                ID_FIELD.to_string(),
                AttributeValue {
                    s: Some(input_id.to_string()),
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

        let update_result = self
            .db_client
            .update_item(UpdateItemInput {
                table_name: "Users".to_string(),
                return_values: Some("ALL_NEW".to_string()),
                key: key,
                attribute_updates: None,
                condition_expression: None,
                conditional_operator: None,
                expected: None,
                expression_attribute_names: None,
                expression_attribute_values: None,
                return_consumed_capacity: None,
                return_item_collection_metrics: None,
                update_expression: None,
            })
            .await?;

        self.extract_id(update_result)
    }

    fn extract_id(&self, output: UpdateItemOutput) -> Result<String, Box<dyn Error>> {
        let attributes = {
            match output.attributes {
                Some(attributes) => attributes,
                None => {
                    return Err(Box::new(DBError {
                        detail: "No attributes found on user update call".to_string(),
                    }))
                }
            }
        };

        let id_value = {
            match attributes.get(ID_FIELD) {
                Some(attribute_value) => attribute_value,
                None => {
                    return Err(Box::new(DBError {
                        detail: "No id found in user attributes".to_string(),
                    }))
                }
            }
        };

        match &id_value.s {
            Some(id_str) => Ok(id_str.to_string()),
            None => {
                return Err(Box::new(DBError {
                    detail: "ID attribute on user is not a string".to_string(),
                }))
            }
        }
    }
}
