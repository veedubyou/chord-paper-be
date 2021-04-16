mod application;
mod db_utils;
mod environment;
mod server;

use application::concerns::dynamodb::deserialize;
use application::concerns::dynamodb::expression;

use rusoto_dynamodb::AttributeValue;
use serde_json;
use std::collections::HashMap;

fn main() {
    testA();
    // server::serve(([0, 0, 0, 0], 5000)).await;
}

use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct Track {
    #[serde(flatten)]
    pub extra: serde_json::Map<String, serde_json::Value>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct TrackList {
    pub tracks: Vec<Track>,
}

fn testA() {
    let track = Track {
        extra: {
            let mut url_field = serde_json::Map::new();
            url_field.insert(
                "vocals".to_string(),
                serde_json::Value::String("http://whateverdude".to_string()),
            );

            let mut vocal_stem = serde_json::Map::new();
            vocal_stem.insert("vocals".to_string(), serde_json::Value::Object(url_field));

            let mut track_stem = serde_json::Map::new();
            track_stem.insert("stems".to_string(), serde_json::Value::Object(vocal_stem));
            track_stem
        },
    };

    let serialized_result = serde_dynamodb::to_hashmap(&track);
    match serialized_result {
        Ok(serialized) => {
            // println!("serialized!");
            // println!("{:#?}", serialized);

            let deserialised: Result<Track, serde_dynamodb::Error> =
                serde_dynamodb::from_hashmap(serialized);
            match deserialised {
                Ok(d) => {
                    println!("deserialized!");
                    println!("{:#?}", d);
                }
                Err(e) => {
                    println!("failed deserialization!");
                    println!("{:#?}", e);
                }
            }
        }
        Err(e) => println!("{:#?}", e),
    }
}

fn test() {
    let stem_url = AttributeValue {
        s: Some("http://localhost:3000/vocals.mp3".to_string()),
        ..Default::default()
    };

    let vocals_stem = AttributeValue {
        m: Some(expression::make_hashmap(
            "url".to_string(),
            stem_url.clone(),
        )),
        ..Default::default()
    };

    let mut stems_map = HashMap::new();
    stems_map.insert("vocals".to_string(), vocals_stem);

    let stems = AttributeValue {
        m: Some(stems_map),
        ..Default::default()
    };

    let mut track_map = HashMap::new();
    track_map.insert("stems".to_string(), stems);
    // track_map.insert("id".to_string(), id);
    // track_map.insert("track_type".to_string(), track_type);

    let track = AttributeValue {
        m: Some(track_map),
        ..Default::default()
    };

    let tracks = AttributeValue {
        l: Some(vec![track]),
        ..Default::default()
    };

    let mut tracklist_map = HashMap::new();
    tracklist_map.insert("tracks".to_string(), tracks);

    println!("{:#?}", tracklist_map);

    let result: Result<TrackList, serde_dynamodb::Error> =
        serde_dynamodb::from_hashmap(tracklist_map);
    match result {
        Ok(d) => println!("{:#?}", d),
        Err(e) => println!("{:#?}", e),
    }
}
