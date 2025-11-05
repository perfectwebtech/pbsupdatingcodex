use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("Database error: {0}")]
    Database(#[from] tokio_postgres::Error),

    #[error("Redis error: {0}")]
    Redis(#[from] redis::RedisError),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("JWT error: {0}")]
    Jwt(#[from] jsonwebtoken::errors::Error),

    #[error("Job not found")]
    JobNotFound,

    #[error("Invalid job state: {0}")]
    InvalidJobState(String),

    #[error("Unauthorized")]
    Unauthorized,

    #[error("Internal server error")]
    Internal,
}

impl actix_web::error::ResponseError for Error {
    fn error_response(&self) -> actix_web::HttpResponse {
        match self {
            Error::JobNotFound => actix_web::HttpResponse::NotFound().json(
                serde_json::json!({"success": false, "error": self.to_string()})
            ),
            Error::Unauthorized => actix_web::HttpResponse::Unauthorized().json(
                serde_json::json!({"success": false, "error": self.to_string()})
            ),
            _ => actix_web::HttpResponse::InternalServerError().json(
                serde_json::json!({"success": false, "error": "Internal server error"})
            ),
        }
    }
}

pub type Result<T> = std::result::Result<T, Error>;
