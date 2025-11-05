use serde::Deserialize;
use config::{ConfigError, Environment};

#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    pub server: ServerConfig,
    pub database: DatabaseConfig,
    pub redis: RedisConfig,
    pub transcoding: TranscodingConfig,
    pub jwt: JwtConfig,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ServerConfig {
    pub port: u16,
    pub workers: usize,
}

#[derive(Debug, Clone, Deserialize)]
pub struct DatabaseConfig {
    pub host: String,
    pub port: u16,
    pub username: String,
    pub password: String,
    pub database: String,
    pub max_connections: u32,
}

#[derive(Debug, Clone, Deserialize)]
pub struct RedisConfig {
    pub host: String,
    pub port: u16,
    pub password: Option<String>,
    pub db: u8,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TranscodingConfig {
    pub max_concurrent_jobs: usize,
    pub input_path: String,
    pub output_path: String,
    pub temp_path: String,
    pub presets: Vec<TranscodingPreset>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TranscodingPreset {
    pub name: String,
    pub video_codec: String,
    pub video_bitrate: String,
    pub audio_codec: String,
    pub audio_bitrate: String,
    pub resolution: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct JwtConfig {
    pub secret: String,
}

impl Config {
    pub fn from_env() -> Result<Self, ConfigError> {
        dotenv::dotenv().ok();

        let mut builder = config::Config::builder()
            .set_default("server.port", 8002)?
            .set_default("server.workers", 4)?
            .set_default("database.host", "localhost")?
            .set_default("database.port", 5432)?
            .set_default("database.max_connections", 16)?
            .set_default("redis.host", "localhost")?
            .set_default("redis.port", 6379)?
            .set_default("redis.db", 0)?
            .set_default("transcoding.max_concurrent_jobs", 10)?
            .set_default("transcoding.input_path", "/data/input")?
            .set_default("transcoding.output_path", "/data/output")?
            .set_default("transcoding.temp_path", "/data/temp")?;

        // Add environment variables
        builder = builder.add_source(Environment::default().separator("__"));

        builder.build()?.try_deserialize()
    }
}
