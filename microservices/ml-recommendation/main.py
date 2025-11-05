import uvicorn
from fastapi import FastAPI, Depends, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager

from app.config import settings
from app.routers import recommendations, training
from app.middleware.auth import JWTBearer
from app.database import engine, Base
from app.models.ml import RecommendationModel

# Global model instance
ml_model = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources"""
    global ml_model

    # Startup
    print("Loading ML model...")
    ml_model = RecommendationModel()
    await ml_model.load_model()
    print("ML model loaded successfully")

    # Create database tables
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

    yield

    # Shutdown
    print("Shutting down...")

app = FastAPI(
    title="ML Recommendation Service",
    description="AI-powered content recommendation engine",
    version="1.0.0",
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Health check
@app.get("/health")
async def health_check():
    return {
        "status": "ok",
        "service": "ml-recommendation",
        "model_loaded": ml_model is not None
    }

# Routers
app.include_router(
    recommendations.router,
    prefix="/api/v1/recommendations",
    tags=["recommendations"],
    dependencies=[Depends(JWTBearer())]
)

app.include_router(
    training.router,
    prefix="/api/v1/training",
    tags=["training"],
    dependencies=[Depends(JWTBearer())]
)

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.PORT,
        reload=settings.DEBUG,
        workers=settings.WORKERS
    )
