# Video Recommendation System

## Overview

The Analytics Service includes a comprehensive recommendation system that provides personalized
video recommendations using both content-based and collaborative filtering approaches. The system is
designed to improve user engagement by suggesting relevant videos based on user preferences and
behavior patterns.

## Architecture

### Components

1. **ContentBasedRecommender**: Uses video features (tags, quality metrics, metadata) to find
   similar content
2. **CollaborativeFilteringRecommender**: Uses user interaction patterns to find similar users and
   recommend videos
3. **HybridRecommendationEngine**: Combines both approaches for better recommendations
4. **RecommendationCache**: Redis-based caching for improved performance

### Features

- **Content-Based Filtering**: Analyzes video features like tags, quality scores, duration,
  resolution
- **Collaborative Filtering**: Uses SVD matrix factorization on user-item interaction matrix
- **Hybrid Approach**: Combines both methods with configurable weights (60% content, 40%
  collaborative)
- **Real-time Caching**: Redis caching with automatic invalidation
- **User Preference Tracking**: Learns from user interactions (views, downloads, likes, shares)
- **Automatic Model Retraining**: Periodic retraining with new data

## API Endpoints

### Get User Recommendations

```http
GET /api/v1/recommendations/user/{user_id}?limit=10&use_cache=true
```

Returns personalized recommendations for a user based on their preferences and behavior.

**Response:**

```json
[
  {
    "video_id": "video123",
    "score": 0.85,
    "title": "Amazing Guitar Solo",
    "tags": ["music", "rock", "guitar"],
    "quality_score": 0.9,
    "thumbnails": ["thumb1.jpg", "thumb2.jpg"],
    "duration": 180,
    "user_id": "user456",
    "created_at": "2023-10-01T12:00:00Z"
  }
]
```

### Get Similar Videos

```http
GET /api/v1/recommendations/video/{video_id}/similar?limit=10&use_cache=true
```

Returns videos similar to a specific video based on content features.

### Track User Interaction

```http
POST /api/v1/recommendations/user/{user_id}/interaction
```

**Request Body:**

```json
{
  "video_id": "video123",
  "interaction_type": "view" // view, download, like, share
}
```

Tracks user interactions to improve future recommendations.

### Retrain Models

```http
GET /api/v1/recommendations/retrain
```

Manually triggers retraining of recommendation models with latest data.

### Health Check

```http
GET /api/v1/recommendations/health
```

Returns the health status of the recommendation system.

## Data Models

### User Preferences

```python
{
  "user_id": "user123",
  "preferred_tags": {"music": 5, "rock": 3, "guitar": 2},
  "quality_preference": 0.8,
  "interaction_count": 25,
  "interactions": [
    {
      "video_id": "video123",
      "type": "view",
      "timestamp": "2023-10-01T12:00:00Z"
    }
  ]
}
```

### Video Features

```python
{
  "video_id": "video123",
  "user_id": "user456",
  "features": {
    "quality_score": 0.8,
    "sharpness": 0.9,
    "brightness": 0.7,
    "contrast": 0.8,
    "duration": 180,
    "resolution_width": 1920,
    "resolution_height": 1080,
    "aspect_ratio": 1.77,
    "fps": 30,
    "safety_score": 0.9,
    "tag_music": 1,
    "tag_rock": 1,
    "tag_guitar": 1
  }
}
```

## Algorithm Details

### Content-Based Filtering

1. **Feature Extraction**: Combines text features (tags, titles) using TF-IDF vectorization and
   numerical features (quality metrics, metadata)
2. **Similarity Calculation**: Uses cosine similarity to find similar videos
3. **User Profile Building**: Creates user preference vectors based on interaction history
4. **Recommendation Generation**: Finds videos most similar to user preferences

### Collaborative Filtering

1. **User-Item Matrix**: Builds matrix of user interactions with videos
2. **Interaction Weighting**: Different weights for different interaction types:
   - View: 1.0
   - Download: 2.0
   - Like: 3.0
   - Share: 2.5
3. **Matrix Factorization**: Uses Truncated SVD for dimensionality reduction
4. **Prediction**: Predicts user ratings for unseen videos

### Hybrid Approach

Combines both methods using weighted scoring:

- Content-based weight: 60%
- Collaborative filtering weight: 40%

## Performance Optimizations

### Caching Strategy

- **User Recommendations**: Cached for 1 hour
- **Similar Videos**: Cached for 1 hour
- **Cache Invalidation**: Automatic invalidation when user interactions change

### Model Training

- **Periodic Retraining**: Every 6 hours
- **Incremental Updates**: New data triggers background retraining
- **Feature Matrix Optimization**: Efficient sparse matrix operations

### Database Indexing

- User ID and video ID indexes for fast lookups
- Timestamp indexes for recent interaction queries
- Tag indexes for content-based filtering

## Configuration

### Environment Variables

```bash
REDIS_URL=redis://localhost:6379
MONGODB_URL=mongodb://localhost:27017
MONGODB_DATABASE=analytics
```

### Model Parameters

```python
# Content-based recommender
TF_IDF_MAX_FEATURES = 1000
MIN_SIMILARITY_THRESHOLD = 0.1

# Collaborative filtering
SVD_COMPONENTS = 50  # Auto-adjusted based on data size
INTERACTION_WEIGHTS = {
    'view': 1.0,
    'download': 2.0,
    'like': 3.0,
    'share': 2.5
}

# Hybrid weights
CONTENT_WEIGHT = 0.6
COLLABORATIVE_WEIGHT = 0.4

# Caching
CACHE_TTL = 3600  # 1 hour
TRAINING_INTERVAL = 6 * 3600  # 6 hours
```

## Usage Examples

### Initialize Recommendation Engine

```python
from analytics.recommendation_engine import HybridRecommendationEngine

engine = HybridRecommendationEngine()
await engine.initialize()
```

### Get Recommendations

```python
recommendations = await engine.get_recommendations(
    user_id="user123",
    n_recommendations=10,
    use_cache=True
)
```

### Track User Interaction

```python
await engine.update_user_interaction(
    user_id="user123",
    video_id="video456",
    interaction_type="like"
)
```

### Train Models

```python
await engine.train_models(force_retrain=True)
```

## Monitoring and Metrics

### Health Metrics

- Model training status
- Cache connection status
- Feature matrix dimensions
- User-item matrix size
- Last training timestamp

### Performance Metrics

- Recommendation response time
- Cache hit rate
- Model accuracy (if ground truth available)
- User engagement metrics

## Future Enhancements

1. **Deep Learning Models**: Integration with neural collaborative filtering
2. **Real-time Learning**: Online learning algorithms for immediate adaptation
3. **Multi-armed Bandits**: Exploration vs exploitation for new content
4. **Contextual Recommendations**: Time-based and location-based recommendations
5. **A/B Testing Framework**: Built-in experimentation capabilities
6. **Diversity Optimization**: Ensuring recommendation diversity
7. **Cold Start Solutions**: Better handling of new users and videos

## Testing

Run the test suite:

```bash
cd services/analytics
uv run python test_recommendation_system.py
```

The test covers:

- Content-based recommendation generation
- Collaborative filtering with user-item matrix
- Feature extraction and similarity calculation
- User preference modeling

## Troubleshooting

### Common Issues

1. **SVD Components Error**: Automatically adjusts based on data size
2. **Cache Connection Issues**: Falls back to direct computation
3. **Empty Recommendations**: Returns popular videos as fallback
4. **Model Training Failures**: Logs errors and continues with existing models

### Debug Mode

Enable debug logging to see detailed recommendation process:

```python
import logging
logging.getLogger('analytics.recommendation_engine').setLevel(logging.DEBUG)
```
