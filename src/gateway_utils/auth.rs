use std::error::Error;
use std::fmt;

#[derive(Debug, Clone)]
struct AuthError;

impl fmt::Display for AuthError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "Unauthenticated error")
    }
}

impl Error for AuthError {}

pub fn extract_bearer_token(auth_header_value: &str) -> Result<String, Box<dyn Error>> {
    let strip_result = auth_header_value.strip_prefix("Bearer ");
    match strip_result {
        Some(auth_token) => Ok(auth_token.to_owned()),
        None => Err(Box::new(AuthError{}))
    }
}