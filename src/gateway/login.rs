use super::auth;
use crate::usecases;
use crate::usecases::login::User;
use http::StatusCode;

pub struct Google {
    usecase: usecases::login::Google,
}

impl Google {
    pub fn new(google_client_id: &str) -> Google {
        Google {
            usecase: usecases::login::Google::new(google_client_id),
        }
    }

    pub fn verify_login(&self, user_id: &str, auth_header_value: &str) -> Box<dyn warp::Reply> {
        let token: String = match auth::extract_bearer_token(&auth_header_value) {
            Ok(auth_token) => auth_token,
            Err(error) => {
                return Box::new( warp::reply::with_status(error.to_string(), StatusCode::UNAUTHORIZED))
            }
        };

        let user: User = match self.usecase.verify_login(&token) {
            Ok(user) => user,
            Err(error) => {
                return Box::new( warp::reply::with_status(error.to_string(), StatusCode::UNAUTHORIZED))
            },
        };

        if user.id != user_id {
            return Box::new( warp::reply::with_status("Requested user ID does not match id token", StatusCode::UNAUTHORIZED))
        }

        Box::new(warp::reply::json(&user))
    }
}