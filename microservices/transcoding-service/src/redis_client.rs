use redis::Client;
use crate::config::RedisConfig;

pub async fn create_client(config: &RedisConfig) -> Result<Client, redis::RedisError> {
    let connection_string = if let Some(password) = &config.password {
        format!("redis://:{}@{}:{}/{}", password, config.host, config.port, config.db)
    } else {
        format!("redis://{}:{}/{}", config.host, config.port, config.db)
    };

    Client::open(connection_string)
}
