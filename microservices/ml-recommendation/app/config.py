from pydantic_settings import BaseSettings
from functools import lru_cache

class Settings(BaseSettings):
    # Server
    PORT: int = 8003
    WORKERS: int = 2
    DEBUG: bool = False

    # Database
    DATABASE_URL: str = "postgresql+asyncpg://iptv_user:password@postgres:5432/iptv_platform"

    # Redis
    REDIS_HOST: str = "redis"
    REDIS_PORT: int = 6379
    REDIS_DB: int = 0
    REDIS_PASSWORD: str = ""

    # JWT
    JWT_SECRET: str = "change_me_in_production"
    JWT_ALGORITHM: str = "HS256"

    # ML Model
    MODEL_PATH: str = "/app/models"
    BATCH_SIZE: int = 32
    MAX_RECOMMENDATIONS: int = 50

    # Training
    TRAINING_ENABLED: bool = True
    TRAINING_INTERVAL_HOURS: int = 24

    class Config:
        env_file = ".env"
        case_sensitive = True

@lru_cache()
def get_settings():
    return Settings()

settings = get_settings()
