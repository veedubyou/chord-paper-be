use super::entity;
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

        let user: entity::User = match self.usecase.login(&token).await {
            Ok(user) => user,
            Err(err) => {
                let status_code = match err {
                    usecase::Error::VerificationError { .. } => StatusCode::UNAUTHORIZED,
                    usecase::Error::DatastoreError { .. } => StatusCode::INTERNAL_SERVER_ERROR,
                };

                return Box::new(warp::reply::with_status(err.to_string(), status_code));
            }
        };

        Box::new(warp::reply::json(&user))
    }
}
