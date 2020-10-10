use serde::{Deserialize, Serialize};

// this struct is entirely structured around the JSON representation being the canonical
// since the primary purpose is to persist and retrieve this data
// capturing it as a grab bag of keys/values makes it very difficult to drop any attributes when persisting
// as opposed to if the struct was thoroughly typed
#[derive(Serialize, Deserialize, Debug)]
pub struct Song {
    pub id: String,
    pub owner: String,
    elements: Vec<serde_json::Value>,
    metadata: serde_json::Map<String, serde_json::Value>,
    #[serde(flatten)]
    extra: serde_json::Map<String, serde_json::Value>,
}

impl Song {
    pub fn is_new(&self) -> bool {
        self.id.is_empty()
    }

    pub fn create_id(&mut self) {
        if !self.is_new() {
            panic!("Cannot assign an ID to a song that already has one")
        }

        self.id = uuid::Uuid::new_v4().to_string()
    }
}
