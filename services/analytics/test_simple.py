#!/usr/bin/env python3
"""
Simple test to verify the quality analysis modules can be imported and basic functionality works
"""
import sys
import os

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

def test_imports():
    """Test that all modules can be imported"""
    print("Testing imports...")
    
    try:
        from analytics.quality_analyzer import VideoQualityAnalyzer, ContentModerationAnalyzer, QualityScoringAlgorithm
        print("✓ Successfully imported quality analyzer classes")
        
        from analytics.quality_analyzer import quality_analyzer, content_moderator, quality_scorer
        print("✓ Successfully imported global instances")
        
        # Test that instances are of correct types
        assert isinstance(quality_analyzer, VideoQualityAnalyzer)
        assert isinstance(content_moderator, ContentModerationAnalyzer)
        assert isinstance(quality_scorer, QualityScoringAlgorithm)
        print("✓ Global instances are of correct types")
        
        return True
        
    except ImportError as e:
        print(f"❌ Import failed: {e}")
        return False
    except Exception as e:
        print(f"❌ Unexpected error: {e}")
        return False


def test_quality_scorer():
    """Test the quality scoring algorithm with mock data"""
    print("\nTesting quality scoring algorithm...")
    
    try:
        from analytics.quality_analyzer import quality_scorer
        
        # Mock data
        quality_metrics = {
            "sharpness_score": 0.8,
            "brightness_score": 0.7,
            "contrast_score": 0.9,
            "noise_score": 0.6
        }
        
        safety_metrics = {
            "overall_score": 0.9
        }
        
        metadata = {
            "width": 1920,
            "height": 1080,
            "fps": 30
        }
        
        # Test scoring
        result = quality_scorer.calculate_weighted_score(quality_metrics, safety_metrics, metadata)
        
        # Verify result structure
        expected_keys = [
            "overall_score", "technical_quality", "visual_appeal", 
            "content_quality", "safety_contribution", "quality_category"
        ]
        
        for key in expected_keys:
            assert key in result, f"Missing key: {key}"
        
        # Verify score ranges
        assert 0.0 <= result["overall_score"] <= 1.0, "Overall score out of range"
        assert result["quality_category"] in [
            "exceptional", "excellent", "very_good", "good", "fair", "poor", "very_poor"
        ], "Invalid quality category"
        
        print(f"✓ Quality scoring test passed - Overall Score: {result['overall_score']:.3f}")
        return True
        
    except Exception as e:
        print(f"❌ Quality scoring test failed: {e}")
        import traceback
        traceback.print_exc()
        return False


def test_content_moderation_heuristics():
    """Test content moderation heuristics without video files"""
    print("\nTesting content moderation heuristics...")
    
    try:
        from analytics.quality_analyzer import content_moderator
        
        # Test heuristic analysis with mock metadata
        metadata = {
            "duration": 300,  # 5 minutes
            "width": 1920,
            "height": 1080,
            "file_size": 50 * 1024 * 1024,  # 50MB
            "detailed": {
                "format": {
                    "tags": {
                        "title": "Safe Test Video",
                        "comment": "Educational content"
                    }
                }
            }
        }
        
        # Test heuristic analysis
        result = content_moderator._analyze_content_heuristics(metadata)
        
        # Verify result structure
        assert "score" in result
        assert "flags" in result
        assert "method" in result
        assert 0.0 <= result["score"] <= 1.0
        
        print(f"✓ Content heuristics test passed - Safety Score: {result['score']:.3f}")
        return True
        
    except Exception as e:
        print(f"❌ Content moderation test failed: {e}")
        import traceback
        traceback.print_exc()
        return False


def main():
    """Run all tests"""
    print("Starting Simple Quality Analysis Tests")
    print("=" * 50)
    
    tests_passed = 0
    total_tests = 3
    
    # Test imports
    if test_imports():
        tests_passed += 1
    
    # Test quality scoring
    if test_quality_scorer():
        tests_passed += 1
    
    # Test content moderation heuristics
    if test_content_moderation_heuristics():
        tests_passed += 1
    
    print("\n" + "=" * 50)
    print(f"Tests passed: {tests_passed}/{total_tests}")
    
    if tests_passed == total_tests:
        print("All tests passed successfully! ✓")
        return 0
    else:
        print("Some tests failed! ❌")
        return 1


if __name__ == "__main__":
    sys.exit(main())