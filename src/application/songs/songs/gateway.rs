use super::entity;
use super::usecase;
use crate::application::concerns::gateway::auth;
use crate::application::concerns::gateway::errors::{
    error_reply, BadRequestError, ForbiddenError, GatewayError, InternalServerError, NotFoundError,
    UnauthorizedError,
};

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

    pub async fn update_song(
        &self,
        auth_header_value: &str,
        song_id: &str,
        song: entity::Song,
    ) -> Box<dyn warp::Reply> {
        let token: String = match auth::extract_auth_token(&auth_header_value) {
            Ok(token) => token,
            Err(reply) => return reply,
        };

        let update_song_result = self.usecase.update_song(&token, song_id, song).await;

        match update_song_result {
            Ok(song) => Box::new(warp::reply::json(&song)),
            Err(err) => map_usecase_errors(err),
        }
    }

    pub async fn delete_song(
        &self,
        auth_header_value: &str,
        song_id: &str,
    ) -> Box<dyn warp::Reply> {
        let token: String = match auth::extract_auth_token(&auth_header_value) {
            Ok(token) => token,
            Err(reply) => return reply,
        };

        let delete_song_result = self.usecase.delete_song(&token, song_id).await;

        match delete_song_result {
            Ok(()) => Box::new(warp::reply()),
            Err(err) => map_usecase_errors(err),
        }
    }
}

pub fn map_usecase_errors(err: usecase::Error) -> Box<dyn warp::Reply> {
    let gateway_error: Box<dyn GatewayError> = match err {
        usecase::Error::GoogleVerificationError { source } => {
            Box::new(UnauthorizedError::FailedGoogleVerification {
                msg: source.to_string(),
            })
        }
        usecase::Error::WrongOwnerError => Box::new(ForbiddenError::UpdateSongOwnerNotAllowed {
            msg: "You do not have permission to modify this user's songs".to_string(),
        }),
        usecase::Error::WrongIDError {
            song_id_1,
            song_id_2,
        } => Box::new(ForbiddenError::UpdateSongWrongId {
            msg: format!(
                "The requested resource ID and the payload ID do not match: {}, {}",
                song_id_1, song_id_2
            ),
        }),
        usecase::Error::ExistingSongError => Box::new(BadRequestError::CreateSongExists {
            msg: "Cannot create a song that already exists".to_string(),
        }),
        usecase::Error::OverwriteError => Box::new(BadRequestError::UpdateSongOverwrite {
            msg: "Cannot update this song - contents will be clobbered. Please refresh the song and try again".to_string(),
        }),
        usecase::Error::NotFoundError { id } => Box::new(NotFoundError::SongNotFound {
            msg: format!("Song ID not found: {}", id),
        }),
        usecase::Error::DatastoreError { source } => Box::new(InternalServerError::DatastoreError {
            msg: source.to_string(),
        }),
    };

    error_reply(gateway_error)
}
