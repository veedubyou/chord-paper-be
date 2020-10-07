use std::env;

#[derive(PartialEq)]
pub enum Environment {
    Development,
    Production,
}

pub fn get_environment() -> Environment {
    let env_var_result = env::var("ENVIRONMENT");

    let environment_result = env_var_result.map(|env_var_str: String| -> Environment {
        match env_var_str.as_str() {
            "development" => Environment::Development,
            "production" => Environment::Production,
            _ => panic!("Invalid environment set, environment: {}", env_var_str),
        }
    });

    match environment_result {
        Ok(environment) => environment,
        Err(err) => panic!("No environment set, error: {}", err),
    }
}
