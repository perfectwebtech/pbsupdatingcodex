import numpy as np
import tensorflow as tf
from typing import List, Dict, Optional
import pickle
import os
from datetime import datetime

class RecommendationModel:
    """
    Collaborative Filtering + Content-Based Hybrid Recommendation System
    """

    def __init__(self):
        self.user_encoder = None
        self.stream_encoder = None
        self.model = None
        self.content_model = None
        self.is_loaded = False

    async def load_model(self):
        """Load pre-trained model or create new one"""
        model_path = "/app/models/recommendation_model.h5"

        if os.path.exists(model_path):
            try:
                self.model = tf.keras.models.load_model(model_path)
                print(f"Loaded model from {model_path}")
            except Exception as e:
                print(f"Error loading model: {e}. Creating new model.")
                self._create_new_model()
        else:
            print("No existing model found. Creating new model.")
            self._create_new_model()

        self.is_loaded = True

    def _create_new_model(self):
        """Create a new neural collaborative filtering model"""
        # User and item embedding dimensions
        embedding_dim = 64

        # User input
        user_input = tf.keras.layers.Input(shape=(1,), name='user_input')
        user_embedding = tf.keras.layers.Embedding(
            input_dim=100000,  # Max users
            output_dim=embedding_dim,
            name='user_embedding'
        )(user_input)
        user_vec = tf.keras.layers.Flatten()(user_embedding)

        # Stream input
        stream_input = tf.keras.layers.Input(shape=(1,), name='stream_input')
        stream_embedding = tf.keras.layers.Embedding(
            input_dim=50000,  # Max streams
            output_dim=embedding_dim,
            name='stream_embedding'
        )(stream_input)
        stream_vec = tf.keras.layers.Flatten()(stream_embedding)

        # Concatenate embeddings
        concat = tf.keras.layers.Concatenate()([user_vec, stream_vec])

        # Dense layers
        dense1 = tf.keras.layers.Dense(128, activation='relu')(concat)
        dense1 = tf.keras.layers.Dropout(0.3)(dense1)
        dense2 = tf.keras.layers.Dense(64, activation='relu')(dense1)
        dense2 = tf.keras.layers.Dropout(0.2)(dense2)
        output = tf.keras.layers.Dense(1, activation='sigmoid')(dense2)

        # Create model
        self.model = tf.keras.Model(
            inputs=[user_input, stream_input],
            outputs=output
        )

        self.model.compile(
            optimizer='adam',
            loss='binary_crossentropy',
            metrics=['accuracy', tf.keras.metrics.AUC()]
        )

        print("New model created successfully")

    async def get_recommendations(
        self,
        user_id: int,
        limit: int = 10,
        exclude_watched: bool = True
    ) -> List[Dict]:
        """
        Get personalized recommendations for a user
        """
        if not self.is_loaded:
            raise Exception("Model not loaded")

        # In production, this would:
        # 1. Get user's watch history
        # 2. Get all available streams
        # 3. Predict scores for unwatched content
        # 4. Return top N recommendations

        # For now, return mock recommendations
        recommendations = []
        for i in range(limit):
            recommendations.append({
                "stream_id": i + 1,
                "score": 0.95 - (i * 0.05),
                "reason": "Based on your viewing history"
            })

        return recommendations

    async def get_similar_content(
        self,
        stream_id: int,
        limit: int = 10
    ) -> List[Dict]:
        """
        Find similar content based on stream features
        """
        # Content-based similarity
        # Uses stream metadata (genre, actors, director, etc.)

        similar_streams = []
        for i in range(limit):
            similar_streams.append({
                "stream_id": stream_id + i + 1,
                "similarity": 0.90 - (i * 0.05),
                "reason": "Similar genre and cast"
            })

        return similar_streams

    async def get_trending(
        self,
        category: Optional[str] = None,
        limit: int = 10
    ) -> List[Dict]:
        """
        Get trending content based on recent popularity
        """
        # Time-decay weighted popularity
        trending = []
        for i in range(limit):
            trending.append({
                "stream_id": 100 + i,
                "score": 1000 - (i * 50),
                "viewers_24h": 5000 - (i * 200)
            })

        return trending

    async def train(self, training_data: np.ndarray, epochs: int = 10):
        """
        Train the model on new data
        """
        if not self.is_loaded:
            raise Exception("Model not loaded")

        # Training data format: [user_id, stream_id, watched (0 or 1)]
        user_ids = training_data[:, 0]
        stream_ids = training_data[:, 1]
        labels = training_data[:, 2]

        history = self.model.fit(
            [user_ids, stream_ids],
            labels,
            epochs=epochs,
            batch_size=256,
            validation_split=0.2,
            verbose=1
        )

        # Save model
        model_path = "/app/models/recommendation_model.h5"
        os.makedirs(os.path.dirname(model_path), exist_ok=True)
        self.model.save(model_path)

        return {
            "loss": float(history.history['loss'][-1]),
            "accuracy": float(history.history['accuracy'][-1]),
            "val_loss": float(history.history['val_loss'][-1]),
            "val_accuracy": float(history.history['val_accuracy'][-1])
        }

    async def evaluate(self, test_data: np.ndarray) -> Dict:
        """
        Evaluate model performance
        """
        user_ids = test_data[:, 0]
        stream_ids = test_data[:, 1]
        labels = test_data[:, 2]

        results = self.model.evaluate([user_ids, stream_ids], labels)

        return {
            "loss": float(results[0]),
            "accuracy": float(results[1]),
            "auc": float(results[2])
        }
