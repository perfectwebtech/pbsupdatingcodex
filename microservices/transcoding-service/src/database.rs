use deadpool_postgres::{Config as PoolConfig, Pool, Runtime};
use tokio_postgres::NoTls;
use crate::config::DatabaseConfig;

pub async fn create_pool(config: &DatabaseConfig) -> Result<Pool, Box<dyn std::error::Error>> {
    let mut pg_config = PoolConfig::new();
    pg_config.host = Some(config.host.clone());
    pg_config.port = Some(config.port);
    pg_config.user = Some(config.username.clone());
    pg_config.password = Some(config.password.clone());
    pg_config.dbname = Some(config.database.clone());

    let pool = pg_config.create_pool(Some(Runtime::Tokio1), NoTls)?;

    Ok(pool)
}
