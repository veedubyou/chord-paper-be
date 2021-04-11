use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug)]
pub struct Track {
    pub id: String,
    pub track_type: String,
    #[serde(flatten)]
    contents: serde_json::Map<String, serde_json::Value>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct TrackList {
    pub song_id: String,
    pub tracks: Vec<Track>,
}

impl TrackList {
    pub fn empty(song_id: &str) -> TrackList {
        TrackList {
            song_id: song_id.to_string(),
            tracks: vec![],
        }
    }
}
