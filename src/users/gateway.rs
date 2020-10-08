use super::entity;
use super::usecase;
use crate::gateway_utils::auth;
use http::StatusCode;

#[derive(Clone)]
pub struct Gateway {
    usecase: usecase::Usecase,
}

impl Gateway {
    pub fn new(usecase: usecase::Usecase) -> Gateway {
        Gateway { usecase: usecase }
    }

    pub async fn login(&self, user_id: &str, auth_header_value: &str) -> Box<dyn warp::Reply> {
        let token: String = match auth::extract_bearer_token(&auth_header_value) {
            Ok(auth_token) => auth_token,
            Err(error) => {
                return Box::new(warp::reply::with_status(
                    error.to_string(),
                    StatusCode::UNAUTHORIZED,
                ))
            }
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

        if user.id != user_id {
            return Box::new(warp::reply::with_status(
                "Requested user ID does not match id token",
                StatusCode::UNAUTHORIZED,
            ));
        }

        Box::new(warp::reply::json(&user))
    }
}
