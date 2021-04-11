use super::usecase;
use crate::application::concerns::gateway::errors::{
    error_reply, GatewayError, InternalServerError,
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
}

pub fn map_usecase_errors(err: usecase::Error) -> Box<dyn warp::Reply> {
    let gateway_error: Box<dyn GatewayError> = match err {
        usecase::Error::DatastoreError { source } => {
            Box::new(InternalServerError::DatastoreError {
                msg: source.to_string(),
            })
        }
    };

    error_reply(gateway_error)
}
