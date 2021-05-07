use super::entity;
use super::usecase;
use crate::application::concerns::gateway::auth::extract_auth_token;
use crate::application::concerns::gateway::errors::{
    error_reply, BadRequestError, ForbiddenError, GatewayError, InternalServerError,
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

    pub async fn get_tracklist(&self, song_id: &str) -> Box<dyn warp::Reply> {
        let get_tracklist_result = self.usecase.get_tracklist(song_id).await;

        match get_tracklist_result {
            Ok(tracklist) => Box::new(warp::reply::json(&tracklist)),
            Err(err) => map_usecase_errors(err),
        }
    }

    pub async fn put_tracklist(
        &self,
        auth_header_value: &str,
        song_id: &str,
        tracklist: entity::TrackList,
    ) -> Box<dyn warp::Reply> {
        let token: String = match extract_auth_token(&auth_header_value) {
            Ok(token) => token,
            Err(reply) => return reply,
        };

        let update_tracklist_result = self.usecase.set_tracklist(&token, song_id, tracklist).await;

        match update_tracklist_result {
            Ok(song) => Box::new(warp::reply::json(&song)),
            Err(err) => map_usecase_errors(err),
        }
    }
}

pub fn map_usecase_errors(err: usecase::Error) -> Box<dyn warp::Reply> {
    let gateway_error: Box<dyn GatewayError> = match err {
        usecase::Error::DatastoreError { source } => {
            Box::new(InternalServerError::DatastoreError {
                msg: source.to_string(),
            })
        }
        usecase::Error::PublishError { msg } => Box::new(InternalServerError::PublishQueueError {
            msg: msg.to_string(),
        }),
        usecase::Error::GoogleVerificationError { source } => {
            Box::new(UnauthorizedError::FailedGoogleVerification {
                msg: source.to_string(),
            })
        }
        usecase::Error::WrongOwnerError => Box::new(ForbiddenError::UpdateSongOwnerNotAllowed {
            msg: "You do not have permission to modify this user's songs".to_string(),
        }),
        usecase::Error::InvalidIDError { .. } => Box::new(BadRequestError::InvalidID {
            msg: "Invalid song ID provided".to_string(),
        }),
        usecase::Error::GetSongError { source } => Box::new(InternalServerError::DatastoreError {
            msg: source.to_string(),
        }),
    };

    error_reply(gateway_error)
}
