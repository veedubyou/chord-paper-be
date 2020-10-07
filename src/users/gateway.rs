use super::Usecase;
use super::User;
use crate::gateway_utils::auth;
use http::StatusCode;

#[derive(Clone)]
pub struct Gateway {
    usecase: Usecase,
}

impl Gateway {
    pub fn new(usecase: Usecase) -> Gateway {
        Gateway { usecase: usecase }
    }

    pub fn login(&self, user_id: &str, auth_header_value: &str) -> Box<dyn warp::Reply> {
        let token: String = match auth::extract_bearer_token(&auth_header_value) {
            Ok(auth_token) => auth_token,
            Err(error) => {
                return Box::new(warp::reply::with_status(
                    error.to_string(),
                    StatusCode::UNAUTHORIZED,
                ))
            }
        };

        let user: User = match self.usecase.login(&token) {
            Ok(user) => user,
            Err(error) => {
                return Box::new(warp::reply::with_status(
                    error.to_string(),
                    StatusCode::UNAUTHORIZED,
                ))
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
