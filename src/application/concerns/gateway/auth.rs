pub fn extract_auth_token(auth_header_value: &str) -> Result<String, Box<dyn warp::Reply>> {
    let strip_result = auth_header_value.strip_prefix("Bearer ");
    match strip_result {
        Some(auth_token) => Ok(auth_token.to_string()),
        None => Err(Box::new(warp::reply::with_status(
            "Unauthenticated error".to_string(),
            http::StatusCode::UNAUTHORIZED,
        ))),
    }
}
