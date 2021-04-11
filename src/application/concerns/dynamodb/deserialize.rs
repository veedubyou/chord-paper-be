use rusoto_dynamodb::AttributeValue;
use serde::Deserialize;
use snafu::Snafu;
use std::collections::HashMap;

#[derive(Debug, Snafu)]
pub enum Error {
    #[snafu(display("Item not found"))]
    NotFoundError,

    #[snafu(display(
        "Data from dynamodb cannot be deserialized into the requested shape: {}",
        source
    ))]
    MalformedDataError { source: serde_dynamodb::Error },
}

pub fn deserialize_from_optional_list_attributes<'a, T: Deserialize<'a>>(
    output: Option<Vec<HashMap<String, AttributeValue>>>,
) -> Result<Vec<T>, Error> {
    let maps = output.ok_or(Error::NotFoundError)?;
    let mut list: Vec<T> = vec![];

    for attribute in maps {
        let object: T = deserialize_from_attributes(attribute)?;
        list.push(object);
    }

    Ok(list)
}

pub fn deserialize_from_optional_attributes<'a, T: Deserialize<'a>>(
    output: Option<HashMap<String, AttributeValue>>,
) -> Result<T, Error> {
    let map = output.ok_or(Error::NotFoundError)?;
    deserialize_from_attributes(map)
}

pub fn deserialize_from_attributes<'a, T: Deserialize<'a>>(
    map: HashMap<String, AttributeValue>,
) -> Result<T, Error> {
    let object: T = serde_dynamodb::from_hashmap(map)
        .map_err(|err| Error::MalformedDataError { source: err })?;
    Ok(object)
}
