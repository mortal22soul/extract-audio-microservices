#!/usr/bin/env python3
"""
Comprehensive test for quality analysis and content moderation functionality
"""
import asyncio
import sys
import os

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from analytics.quality_analyzer import quality_analyzer, content_moderator, quality_scorer


async def test_quality_analysis():
    """Test video quality analysis with mock video path"""
    print("Testing video quality analysis...")
    
    # Test with a mock video path (the function will return mock data)
    mock_video_path = "/tmp/test_video.mp4"
    
    try:
        result = await quality_analyzer.analyze_quality(mock_video_path)
        
        print("Quality Analysis Results:")
        for key, value in result.items():
            if isinstance(value, dict):
                print(f"  {key}:")
                for sub_key, sub_value in value.items():
                    print(f"    {sub_key}: {sub_value}")
            else:
                print(f"  {key}: {value}")
        
        # Verify required fields are present
        required_fields = [
            'sharpness_score', 'brightness_score', 'contrast_score', 
            'noise_score', 'overall_score', 'quality_category'
        ]
        
        for field in required_fields:
            assert field in result, f"Missing required field: {field}"
        
        print("✓ Quality analysis test passed!")
        return result
        
    except Exception as e:
        print(f"❌ Quality analysis test failed: {e}")
        import traceback
        traceback.print_exc()
        return None


async def test_content_moderation():
    """Test content moderation analysis"""
    print("\nTesting content moderation...")
    
    mock_video_path = "/tmp/safe_test_video.mp4"
    mock_metadata = {
        "duration": 300,  # 5 minutes
        "width": 1920,
        "height": 1080,
        "fps": 30,
        "file_size": 50 * 1024 * 1024,  # 50MB
        "detailed": {
            "format": {
                "tags": {
                    "title": "Educational Video",
                    "comment": "Safe educational content"
                }
            }
        }
    }
    
    try:
        result = await content_moderator.analyze_content_safety(mock_video_path, mock_metadata)
        
        print("Content Moderation Results:")
        for key, value in result.items():
            if isinstance(value, list):
                print(f"  {key}: {len(value)} items")
                for i, item in enumerate(value[:3]):  # Show first 3 items
                    print(f"    [{i}]: {item}")
            else:
                print(f"  {key}: {value}")
        
        # Verify required fields are present
        required_fields = [
            'overall_score', 'is_safe', 'confidence', 'flags', 
            'analysis_methods', 'moderation_action'
        ]
        
        for field in required_fields:
            assert field in result, f"Missing required field: {field}"
        
        print("✓ Content moderation test passed!")
        return result
        
    except Exception as e:
        print(f"❌ Content moderation test failed: {e}")
        import traceback
        traceback.print_exc()
        return None


def test_weighted_scoring():
    """Test weighted quality scoring"""
    print("\nTesting weighted quality scoring...")
    
    # Mock quality metrics
    quality_metrics = {
        "sharpness_score": 0.8,
        "brightness_score": 0.7,
        "contrast_score": 0.9,
        "noise_score": 0.6
    }
    
    # Mock safety metrics
    safety_metrics = {
        "overall_score": 0.9,
        "is_safe": True
    }
    
    # Mock metadata
    metadata = {
        "width": 1920,
        "height": 1080,
        "fps": 30,
        "duration": 300
    }
    
    try:
        result = quality_scorer.calculate_weighted_score(
            quality_metrics, safety_metrics, metadata
        )
        
        print("Weighted Quality Scoring Results:")
        for key, value in result.items():
            if isinstance(value, dict):
                print(f"  {key}:")
                for sub_key, sub_value in value.items():
                    print(f"    {sub_key}: {sub_value}")
            else:
                print(f"  {key}: {value}")
        
        # Verify required fields are present
        required_fields = [
            'overall_score', 'technical_quality', 'visual_appeal',
            'content_quality', 'safety_contribution', 'quality_category'
        ]
        
        for field in required_fields:
            assert field in result, f"Missing required field: {field}"
        
        # Verify score ranges
        for score_field in ['overall_score', 'technical_quality', 'visual_appeal', 'content_quality']:
            score = result.get(score_field, -1)
            assert 0.0 <= score <= 1.0, f"{score_field} out of range: {score}"
        
        print("✓ Weighted quality scoring test passed!")
        return result
        
    except Exception as e:
        print(f"❌ Weighted quality scoring test failed: {e}")
        import traceback
        traceback.print_exc()
        return None


def test_content_heuristics():
    """Test content moderation heuristics"""
    print("\nTesting content moderation heuristics...")
    
    # Test with different metadata scenarios
    test_cases = [
        {
            "name": "Normal video",
            "metadata": {
                "duration": 300,  # 5 minutes
                "width": 1920,
                "height": 1080,
                "file_size": 50 * 1024 * 1024
            },
            "expected_safe": True
        },
        {
            "name": "Very long video",
            "metadata": {
                "duration": 8000,  # Over 2 hours
                "width": 1920,
                "height": 1080,
                "file_size": 500 * 1024 * 1024
            },
            "expected_safe": True  # Should still be safe but flagged
        },
        {
            "name": "Very short video",
            "metadata": {
                "duration": 5,  # 5 seconds
                "width": 640,
                "height": 480,
                "file_size": 1024 * 1024
            },
            "expected_safe": True  # Should still be safe but flagged
        }
    ]
    
    try:
        for test_case in test_cases:
            print(f"\n  Testing: {test_case['name']}")
            
            result = content_moderator._analyze_content_heuristics(test_case["metadata"])
            
            print(f"    Safety Score: {result.get('score', 0):.3f}")
            print(f"    Flags: {len(result.get('flags', []))}")
            
            # Verify basic structure
            assert 'score' in result
            assert 'flags' in result
            assert 'method' in result
            assert 0.0 <= result['score'] <= 1.0
            
            print(f"    ✓ {test_case['name']} test passed")
        
        print("✓ Content heuristics tests passed!")
        return True
        
    except Exception as e:
        print(f"❌ Content heuristics test failed: {e}")
        import traceback
        traceback.print_exc()
        return False


async def main():
    """Run all comprehensive tests"""
    print("Starting Comprehensive Quality Analysis and Content Moderation Tests")
    print("=" * 70)
    
    tests_passed = 0
    total_tests = 4
    
    # Test quality analysis
    quality_result = await test_quality_analysis()
    if quality_result:
        tests_passed += 1
    
    # Test content moderation
    safety_result = await test_content_moderation()
    if safety_result:
        tests_passed += 1
    
    # Test weighted scoring
    weighted_result = test_weighted_scoring()
    if weighted_result:
        tests_passed += 1
    
    # Test content heuristics
    if test_content_heuristics():
        tests_passed += 1
    
    print("\n" + "=" * 70)
    print(f"Tests passed: {tests_passed}/{total_tests}")
    
    if tests_passed == total_tests:
        print("🎉 All comprehensive tests passed successfully!")
        
        print("\nFinal Summary:")
        if quality_result:
            print(f"  Quality Analysis: {quality_result.get('overall_score', 0):.3f} ({quality_result.get('quality_category', 'unknown')})")
        if safety_result:
            print(f"  Content Safety: {safety_result.get('overall_score', 0):.3f} ({safety_result.get('moderation_action', 'unknown')})")
        if weighted_result:
            print(f"  Weighted Score: {weighted_result.get('overall_score', 0):.3f} ({weighted_result.get('quality_category', 'unknown')})")
        
        return 0
    else:
        print("❌ Some comprehensive tests failed!")
        return 1


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))