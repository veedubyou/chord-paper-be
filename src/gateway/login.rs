use serde::Deserialize;

use crate::usecases;
use http::StatusCode;

pub struct Google {
    usecase: usecases::login::Google,
}

#[derive(Deserialize)]
pub struct VerifyLoginRequest {
    id_token: String,
}

impl Google {
    pub fn new(google_client_id: &str) -> Google {
        Google {
            usecase: usecases::login::Google::new(google_client_id),
        }
    }

    pub fn verify_login(&self, request: VerifyLoginRequest) -> Box<dyn warp::Reply> {
        let result = self.usecase.verify_login(request.id_token.as_str());

        match result {
            Ok(user) => Box::new(warp::reply::json(&user)),
            Err(error) => Box::new( warp::reply::with_status(error.to_string(), StatusCode::UNAUTHORIZED)),
        }
    }
}
