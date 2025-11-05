use actix_web::{web, App, HttpServer, HttpResponse, middleware};
use transcoding_service::{
    config::Config,
    database,
    redis_client,
    handlers,
    middleware::auth,
    transcoder::TranscoderPool,
};
use tracing::{info, error};
use tracing_subscriber;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize logging
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::from_default_env()
                .add_directive(tracing::Level::INFO.into()),
        )
        .init();

    // Load configuration
    let config = Config::from_env().expect("Failed to load configuration");

    info!("Starting Transcoding Service on port {}", config.server.port);

    // Initialize database pool
    let db_pool = database::create_pool(&config.database)
        .await
        .expect("Failed to create database pool");

    // Initialize Redis client
    let redis_client = redis_client::create_client(&config.redis)
        .await
        .expect("Failed to create Redis client");

    // Initialize transcoder pool
    let transcoder_pool = TranscoderPool::new(config.transcoding.max_concurrent_jobs);

    let db_pool_data = web::Data::new(db_pool);
    let redis_data = web::Data::new(redis_client);
    let transcoder_data = web::Data::new(transcoder_pool);
    let config_data = web::Data::new(config.clone());

    // Start HTTP server
    HttpServer::new(move || {
        App::new()
            .app_data(db_pool_data.clone())
            .app_data(redis_data.clone())
            .app_data(transcoder_data.clone())
            .app_data(config_data.clone())
            .wrap(middleware::Logger::default())
            .wrap(middleware::Compress::default())
            .route("/health", web::get().to(health_check))
            .route("/metrics", web::get().to(handlers::metrics))
            .service(
                web::scope("/api/v1")
                    .wrap(auth::JwtAuth)
                    .route("/transcode", web::post().to(handlers::create_transcode_job))
                    .route("/jobs", web::get().to(handlers::list_jobs))
                    .route("/jobs/{id}", web::get().to(handlers::get_job))
                    .route("/jobs/{id}/cancel", web::post().to(handlers::cancel_job))
                    .route("/jobs/{id}/status", web::get().to(handlers::get_job_status))
            )
    })
    .bind(("0.0.0.0", config.server.port))?
    .workers(config.server.workers)
    .run()
    .await
}

async fn health_check() -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({
        "status": "ok",
        "service": "transcoding-service"
    }))
}
