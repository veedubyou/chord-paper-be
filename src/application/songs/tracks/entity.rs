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

    pub fn ensure_track_ids(&mut self) {
        for track in &mut self.tracks {
            if track.is_new() {
                track.create_id();
            }
        }
    }
}
