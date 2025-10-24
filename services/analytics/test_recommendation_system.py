"""
Test script for the recommendation system
"""
import asyncio
import sys
import os

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from analytics.recommendation_engine import (
    ContentBasedRecommender,
    CollaborativeFilteringRecommender
)


async def test_content_based_recommender():
    """Test content-based recommendation system"""
    print("Testing Content-Based Recommender...")
    
    recommender = ContentBasedRecommender()
    
    # Sample video data
    video_data = [
        {
            'video_id': 'video1',
            'tags': ['music', 'rock', 'guitar'],
            'metadata': {
                'title': 'Rock Guitar Solo',
                'description': 'Amazing guitar performance',
                'duration': 180,
                'width': 1920,
                'height': 1080,
                'fps': 30
            },
            'quality': {
                'overall_score': 0.8,
                'sharpness_score': 0.9,
                'brightness_score': 0.7,
                'contrast_score': 0.8
            },
            'safety': {
                'overall_score': 0.9
            }
        },
        {
            'video_id': 'video2',
            'tags': ['music', 'jazz', 'piano'],
            'metadata': {
                'title': 'Jazz Piano Performance',
                'description': 'Smooth jazz piano',
                'duration': 240,
                'width': 1920,
                'height': 1080,
                'fps': 30
            },
            'quality': {
                'overall_score': 0.7,
                'sharpness_score': 0.8,
                'brightness_score': 0.6,
                'contrast_score': 0.7
            },
            'safety': {
                'overall_score': 0.9
            }
        },
        {
            'video_id': 'video3',
            'tags': ['music', 'rock', 'drums'],
            'metadata': {
                'title': 'Rock Drum Solo',
                'description': 'Powerful drum performance',
                'duration': 200,
                'width': 1920,
                'height': 1080,
                'fps': 30
            },
            'quality': {
                'overall_score': 0.85,
                'sharpness_score': 0.9,
                'brightness_score': 0.8,
                'contrast_score': 0.85
            },
            'safety': {
                'overall_score': 0.9
            }
        }
    ]
    
    # Build feature matrix
    await recommender.build_feature_matrix(video_data)
    
    # Get recommendations for video1
    recommendations = recommender.get_content_recommendations('video1', 2)
    
    print(f"Recommendations for video1: {recommendations}")
    
    # Test user preferences
    user_preferences = {
        'preferred_tags': {'music': 5, 'rock': 3, 'guitar': 2},
        'quality_preference': 0.8
    }
    
    user_recommendations = recommender.get_user_content_recommendations(
        user_preferences, [], 2
    )
    
    print(f"User recommendations: {user_recommendations}")
    print("Content-Based Recommender test completed!\n")


async def test_collaborative_filtering():
    """Test collaborative filtering recommendation system"""
    print("Testing Collaborative Filtering Recommender...")
    
    recommender = CollaborativeFilteringRecommender()
    
    # Sample interaction data
    interaction_data = [
        {'user_id': 'user1', 'video_id': 'video1', 'interaction_type': 'view'},
        {'user_id': 'user1', 'video_id': 'video2', 'interaction_type': 'like'},
        {'user_id': 'user2', 'video_id': 'video1', 'interaction_type': 'view'},
        {'user_id': 'user2', 'video_id': 'video3', 'interaction_type': 'download'},
        {'user_id': 'user3', 'video_id': 'video2', 'interaction_type': 'view'},
        {'user_id': 'user3', 'video_id': 'video3', 'interaction_type': 'like'},
    ]
    
    # Build user-item matrix
    await recommender.build_user_item_matrix(interaction_data)
    
    # Get recommendations for user1
    recommendations = recommender.get_collaborative_recommendations('user1', [], 2)
    
    print(f"Collaborative recommendations for user1: {recommendations}")
    
    # Get similar users
    similar_users = recommender.get_similar_users('user1', 2)
    
    print(f"Similar users to user1: {similar_users}")
    print("Collaborative Filtering Recommender test completed!\n")


async def main():
    """Run all tests"""
    print("Starting Recommendation System Tests...\n")
    
    try:
        await test_content_based_recommender()
        await test_collaborative_filtering()
        
        print("All tests completed successfully!")
        
    except Exception as e:
        print(f"Test failed with error: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    asyncio.run(main())