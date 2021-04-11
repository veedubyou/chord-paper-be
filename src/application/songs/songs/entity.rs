use chrono::Timelike;
use serde::{Deserialize, Serialize};
use std::str::FromStr;

// this struct is entirely structured around the JSON representation being the canonical
// since the primary purpose is to persist and retrieve this data
// capturing it as a grab bag of keys/values makes it very difficult to drop any attributes when persisting
// as opposed to if the struct was thoroughly typed
#[derive(Serialize, Deserialize, Debug)]
pub struct Song {
    #[serde(flatten)]
    pub summary: SongSummary,
    elements: Vec<serde_json::Value>,
    #[serde(flatten)]
    extra: serde_json::Map<String, serde_json::Value>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct SongSummary {
    pub id: String,
    pub owner: String,
    #[serde(rename = "lastSavedAt")]
    pub last_saved_at: Option<chrono::DateTime<chrono::Utc>>,
    metadata: serde_json::Map<String, serde_json::Value>,
}

impl Song {
    pub fn is_valid_id(song_id: &str) -> bool {
        match uuid::Uuid::from_str(song_id) {
            Ok(_) => true,
            Err(_) => false,
        }
    }

    pub fn is_new(&self) -> bool {
        self.summary.id.is_empty()
    }

    pub fn create_id(&mut self) {
        if !self.is_new() {
            panic!("Cannot assign an ID to a song that already has one")
        }

        self.summary.id = uuid::Uuid::new_v4().to_string();
    }

    pub fn set_saved_time(&mut self) {
        // truncate nanoseconds because this will be consumed by the browser
        // and browser dates have only millisecond resolution
        //
        // this will result in only second level resolution but will minimize confusion
        self.summary.last_saved_at = chrono::Utc::now().with_nanosecond(0);
    }
}
