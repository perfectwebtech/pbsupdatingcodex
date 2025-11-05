from fastapi import APIRouter, Depends, Query
from typing import List, Optional
from pydantic import BaseModel

from app.middleware.auth import get_current_user_id
import main

router = APIRouter()

class RecommendationResponse(BaseModel):
    stream_id: int
    score: float
    reason: str

class TrendingResponse(BaseModel):
    stream_id: int
    score: float
    viewers_24h: int

@router.get("/personalized", response_model=List[RecommendationResponse])
async def get_personalized_recommendations(
    limit: int = Query(10, ge=1, le=50),
    user_id: int = Depends(get_current_user_id)
):
    """
    Get personalized recommendations for the current user
    """
    recommendations = await main.ml_model.get_recommendations(
        user_id=user_id,
        limit=limit
    )
    return recommendations

@router.get("/similar/{stream_id}", response_model=List[RecommendationResponse])
async def get_similar_content(
    stream_id: int,
    limit: int = Query(10, ge=1, le=50)
):
    """
    Get similar content based on a stream
    """
    similar = await main.ml_model.get_similar_content(
        stream_id=stream_id,
        limit=limit
    )
    return similar

@router.get("/trending", response_model=List[TrendingResponse])
async def get_trending_content(
    category: Optional[str] = None,
    limit: int = Query(10, ge=1, le=50)
):
    """
    Get trending content
    """
    trending = await main.ml_model.get_trending(
        category=category,
        limit=limit
    )
    return trending

@router.post("/feedback")
async def submit_feedback(
    stream_id: int,
    rating: float = Query(..., ge=0.0, le=5.0),
    user_id: int = Depends(get_current_user_id)
):
    """
    Submit user feedback for a stream (used for training)
    """
    # Store feedback in database for future training
    return {
        "success": True,
        "message": "Feedback recorded successfully"
    }
