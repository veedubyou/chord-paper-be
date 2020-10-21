use http::StatusCode;
use serde::{Deserialize, Serialize};

#[typetag::serde(tag = "type")]
pub trait GatewayError {
    fn status_code(&self) -> http::StatusCode;
}

pub fn error_reply(error: Box<dyn GatewayError>) -> Box<dyn warp::Reply> {
    let json = Box::new(warp::reply::json(&error));
    Box::new(warp::reply::with_status(json, error.status_code()))
}

#[derive(Serialize, Deserialize)]
#[serde(tag = "code")]
#[serde(rename_all = "snake_case")]
pub enum BadRequestError {
    CreateSongExists { msg: String },
    UpdateSongOverwrite { msg: String },
}

#[typetag::serde]
impl GatewayError for BadRequestError {
    fn status_code(&self) -> StatusCode {
        http::StatusCode::BAD_REQUEST
    }
}

#[derive(Serialize, Deserialize)]
#[serde(tag = "code")]
#[serde(rename_all = "snake_case")]
pub enum NotFoundError {
    SongNotFound { msg: String },
}

#[typetag::serde]
impl GatewayError for NotFoundError {
    fn status_code(&self) -> StatusCode {
        http::StatusCode::NOT_FOUND
    }
}

#[derive(Serialize, Deserialize)]
#[serde(tag = "code")]
#[serde(rename_all = "snake_case")]
pub enum UnauthorizedError {
    FailedGoogleVerification { msg: String },
}

#[typetag::serde]
impl GatewayError for UnauthorizedError {
    fn status_code(&self) -> StatusCode {
        http::StatusCode::UNAUTHORIZED
    }
}

#[derive(Serialize, Deserialize)]
#[serde(tag = "code")]
#[serde(rename_all = "snake_case")]
pub enum ForbiddenError {
    GetSongsForUserNotAllowed { msg: String },
    UpdateSongOwnerNotAllowed { msg: String },
    UpdateSongWrongID { msg: String },
}

#[typetag::serde]
impl GatewayError for ForbiddenError {
    fn status_code(&self) -> StatusCode {
        http::StatusCode::FORBIDDEN
    }
}

#[derive(Serialize, Deserialize)]
#[serde(tag = "code")]
#[serde(rename_all = "snake_case")]
pub enum InternalServerError {
    DatastoreError { msg: String },
}

#[typetag::serde]
impl GatewayError for InternalServerError {
    fn status_code(&self) -> StatusCode {
        http::StatusCode::INTERNAL_SERVER_ERROR
    }
}
