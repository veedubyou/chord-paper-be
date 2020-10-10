use super::entity;
use super::usecase;
use crate::application::concerns::gateway::auth;

#[derive(Clone)]
pub struct Gateway {
    usecase: usecase::Usecase,
}

impl Gateway {
    pub fn new(usecase: usecase::Usecase) -> Gateway {
        Gateway { usecase: usecase }
    }

    pub async fn get_song(&self, song_id: &str) -> Box<dyn warp::Reply> {
        let get_song_result = self.usecase.get_song(song_id).await;

        match get_song_result {
            Ok(song) => Box::new(warp::reply::json(&song)),
            Err(err) => map_usecase_errors(err),
        }
    }

    pub async fn create_song(
        &self,
        auth_header_value: &str,
        song: entity::Song,
    ) -> Box<dyn warp::Reply> {
        let token: String = match auth::extract_auth_token(&auth_header_value) {
            Ok(token) => token,
            Err(reply) => return reply,
        };

        let create_song_result = self.usecase.create_song(&token, song).await;

        match create_song_result {
            Ok(song) => Box::new(warp::reply::json(&song)),
            Err(err) => map_usecase_errors(err),
        }
    }
}

fn map_usecase_errors(err: usecase::Error) -> Box<dyn warp::Reply> {
    let status_code = match err {
        usecase::Error::VerificationError { .. } => http::StatusCode::UNAUTHORIZED,
        usecase::Error::ExistingSongError | usecase::Error::WrongOwnerError => {
            http::StatusCode::BAD_REQUEST
        }
        usecase::Error::NotFoundError { .. } => http::StatusCode::NOT_FOUND,
        usecase::Error::DatastoreError { .. } => http::StatusCode::INTERNAL_SERVER_ERROR,
    };

    Box::new(warp::reply::with_status(err.to_string(), status_code))
}
