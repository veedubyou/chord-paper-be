use serde::{Deserialize, Serialize};
use serde_json::json;

static SPLIT_TRACK_TYPES: [&'static str; 3] = ["split_2stems", "split_4stems", "split_5stems"];

#[derive(Serialize, Deserialize, Debug)]
pub struct Track {
    pub id: String,
    pub track_type: String,
    #[serde(flatten)]
    pub contents: serde_json::Map<String, serde_json::Value>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct TrackList {
    pub song_id: String,
    pub tracks: Vec<Track>,
}

impl Track {
    pub fn is_new(&self) -> bool {
        self.id.is_empty()
    }

    pub fn create_id(&mut self) {
        if !self.is_new() {
            panic!("Cannot assign an ID to a track that already has one")
        }

        self.id = uuid::Uuid::new_v4().to_string();
    }

    pub fn is_split_request(&self) -> bool {
        for split_track_type in SPLIT_TRACK_TYPES.iter() {
            if self.track_type == *split_track_type {
                return true;
            }
        }
        false
    }

    pub fn init_split_request(&mut self) {
        if !self.is_split_request() {
            panic!("Cannot init a non split request")
        }

        self.contents
            .insert("job_status".to_string(), json!("requested"));

        self.contents.insert(
            "job_status_message".to_string(),
            json!("The splitting job for the audio has been requested"),
        );

        self.contents
            .insert("job_status_debug_log".to_string(), json!(""));

        let initial_progress_percentage = json!(5);
        self.contents
            .insert("job_progress".to_string(), initial_progress_percentage);
    }
}

impl TrackList {
    pub fn has_valid_id(&self) -> bool {
        !self.song_id.is_empty()
    }

    pub fn empty(song_id: &str) -> TrackList {
        TrackList {
            song_id: song_id.to_string(),
            tracks: vec![],
        }
    }
}
