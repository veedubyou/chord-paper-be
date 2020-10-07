use std::env;
use std::net::SocketAddr;

use crate::environment;
use crate::server::users::create_users_gateway;
use warp::Filter;

pub async fn serve(addr: impl Into<SocketAddr> + 'static) {
    let users_gateway = create_users_gateway();

    let login = warp::post()
        .and(warp::path!("login" / String))
        .and(warp::header::<String>("authorization"))
        .map(move |user_id: String, auth_value: String| users_gateway.login(&user_id, &auth_value))
        .with(cors_filter(vec!["POST"]));

    let paths = login.with(warp::log("info"));

    warp::serve(paths).run(addr).await;
}

fn cors_filter(allowed_methods: Vec<&str>) -> warp::filters::cors::Builder {
    let fe_origins: Vec<String> = accepted_fe_origins();
    let fe_origins_ref: Vec<&str> = fe_origins.iter().map(String::as_str).collect();

    warp::cors()
        .allow_origins(fe_origins_ref)
        .allow_methods(allowed_methods)
        .allow_headers(vec!["content-type", "authorization"])
}

fn accepted_fe_origins() -> Vec<String> {
    if environment::get_environment() == environment::Environment::Development {
        return vec!["http://localhost:3000".to_string()];
    }

    // a comma separated list of host origins
    // e.g. ALLOWED_FE_ORIGINS=http://host1.com,https://host2.net
    match env::var("ALLOWED_FE_ORIGINS") {
        Ok(fe_origin) => fe_origin.split(',').map(str::to_owned).collect(),
        Err(e) => panic!("No CORS FE origins set, error: {}", e),
    }
}
