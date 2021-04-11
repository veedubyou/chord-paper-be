use rusoto_dynamodb::AttributeValue;
use std::collections::HashMap;

pub fn make_single_string(key: &str, value: &str) -> HashMap<String, AttributeValue> {
    make_hashmap(
        key.to_string(),
        AttributeValue {
            s: Some(value.to_string()),
            ..Default::default()
        },
    )
}

pub fn make_hashmap<T>(key: String, value: T) -> HashMap<String, T> {
    let mut map: HashMap<String, T> = HashMap::new();
    map.insert(key, value);
    map
}
