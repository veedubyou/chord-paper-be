use std::net::SocketAddr;
use std::env;

use warp::Filter;
use crate::gateway;

const GOOGLE_CLIENT_ID: &str = "650853277550-ta69qbfcvdl6tb5ogtnh2d07ae9rcdlf.apps.googleusercontent.com";

pub async fn serve(addr: impl Into<SocketAddr> + 'static) {
    let verify_login = warp::post()
        .and(warp::path!("login" / String))
        .and(warp::header::<String>("authorization"))
        .map(|user_id: String, auth_value: String| {
            let gateway = gateway::login::Google::new(GOOGLE_CLIENT_ID);
            gateway.verify_login(&user_id, &auth_value)
        })
        .with(cors_filter(vec!["POST"]));

    let paths = verify_login.with(warp::log("info"));

    warp::serve(paths)
        .run(addr)
        .await;
}



fn cors_filter(allowed_methods: Vec<&str>) -> warp::filters::cors::Builder {
    let fe_origins: Vec<String> = accepted_fe_origins();
    let fe_origins_ref: Vec<&str> = fe_origins.iter()
        .map(String::as_str)
        .collect();

    warp::cors()
        .allow_origins(fe_origins_ref)
        .allow_methods(allowed_methods)
        .allow_headers(vec!["content-type", "authorization"])
}

fn accepted_fe_origins() -> Vec<String> {
    // a comma separated list of host origins
    // e.g. ALLOWED_FE_ORIGINS=http://host1.com,https://host2.net
    match env::var("ALLOWED_FE_ORIGINS") {
        Ok(fe_origin) => {
            fe_origin.split(',')
                .map(str::to_owned)
                .collect()
        },
        Err(e) => panic!("No CORS FE origins set, error: {}", e),
    }
}