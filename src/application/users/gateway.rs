use super::usecase;
use crate::application::concerns::gateway::auth;
use http::StatusCode;

#[derive(Clone)]
pub struct Gateway {
    usecase: usecase::Usecase,
}

impl Gateway {
    pub fn new(usecase: usecase::Usecase) -> Gateway {
        Gateway { usecase: usecase }
    }

    pub async fn login(&self, auth_header_value: &str) -> Box<dyn warp::Reply> {
        let token: String = match auth::extract_auth_token(&auth_header_value) {
            Ok(token) => token,
            Err(reply) => return reply,
        };

        match self.usecase.login(&token).await {
            Ok(user) => Box::new(warp::reply::json(&user)),
            Err(err) => map_usecase_errors(err),
        }
    }

    pub async fn songs_for_user(
        &self,
        auth_header_value: &str,
        user_id: &str,
    ) -> Box<dyn warp::Reply> {
        let token: String = match auth::extract_auth_token(&auth_header_value) {
            Ok(token) => token,
            Err(reply) => return reply,
        };

        match self.usecase.songs_for_user(&token, user_id).await {
            Ok(song_summaries) => Box::new(warp::reply::json(&song_summaries)),
            Err(err) => map_usecase_errors(err),
        }
    }
}

fn map_usecase_errors(err: usecase::Error) -> Box<dyn warp::Reply> {
    let status_code = match err {
        usecase::Error::GoogleVerificationError { .. }
        | usecase::Error::OwnerVerificationError { .. } => StatusCode::UNAUTHORIZED,
        usecase::Error::DatastoreError { .. } => StatusCode::INTERNAL_SERVER_ERROR,
    };

    Box::new(warp::reply::with_status(err.to_string(), status_code))
}
