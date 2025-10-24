#!/usr/bin/env python3
"""
Simple test script to verify quality analysis and content moderation functionality
"""
import asyncio
import sys
import os
import tempfile
import numpy as np
import cv2

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from analytics.quality_analyzer import quality_analyzer, content_moderator, quality_scorer


def create_test_video(filename: str, duration_seconds: int = 5) -> str:
    """Create a simple test video file"""
    # Video properties
    fps = 30
    width, height = 640, 480
    total_frames = duration_seconds * fps
    
    # Create video writer
    fourcc = cv2.VideoWriter_fourcc(*'mp4v')
    out = cv2.VideoWriter(filename, fourcc, fps, (width, height))
    
    # Generate frames with different patterns
    for frame_num in range(total_frames):
        # Create a frame with varying patterns
        frame = np.zeros((height, width, 3), dtype=np.uint8)
        
        # Add some patterns to make it interesting
        if frame_num < total_frames // 3:
            # First third: gradient
            for y in range(height):
                frame[y, :] = [int(255 * y / height), 100, 150]
        elif frame_num < 2 * total_frames // 3:
            # Second third: checkerboard
            for y in range(0, height, 20):
                for x in range(0, width, 20):
                    if (x // 20 + y // 20) % 2 == 0:
                        frame[y:y+20, x:x+20] = [255, 255, 255]
        else:
            # Last third: noise
            frame = np.random.randint(0, 256, (height, width, 3), dtype=np.uint8)
        
        out.write(frame)
    
    out.release()
    return filename


async def test_quality_analysis():
    """Test video quality analysis"""
    print("Testing video quality analysis...")
    
    # Create a temporary test video
    with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as tmp_file:
        test_video_path = tmp_file.name
    
    try:
        # Create test video
        create_test_video(test_video_path)
        print(f"Created test video: {test_video_path}")
        
        # Analyze quality
        quality_result = await quality_analyzer.analyze_quality(test_video_path)
        
        print("Quality Analysis Results:")
        print(f"  Sharpness Score: {quality_result.get('sharpness_score', 0):.3f}")
        print(f"  Brightness Score: {quality_result.get('brightness_score', 0):.3f}")
        print(f"  Contrast Score: {quality_result.get('contrast_score', 0):.3f}")
        print(f"  Noise Score: {quality_result.get('noise_score', 0):.3f}")
        print(f"  Overall Score: {quality_result.get('overall_score', 0):.3f}")
        print(f"  Quality Category: {quality_result.get('quality_category', 'unknown')}")
        print(f"  Resolution Category: {quality_result.get('resolution_category', 'unknown')}")
        print(f"  Frames Analyzed: {quality_result.get('frame_count_analyzed', 0)}")
        
        # Check if we got reasonable results
        assert 'sharpness_score' in quality_result
        assert 'brightness_score' in quality_result
        assert 'contrast_score' in quality_result
        assert 'noise_score' in quality_result
        assert 'overall_score' in quality_result
        assert quality_result['frame_count_analyzed'] > 0
        
        print("✓ Quality analysis test passed!")
        return quality_result
        
    finally:
        # Clean up
        if os.path.exists(test_video_path):
            os.unlink(test_video_path)


async def test_content_moderation():
    """Test content moderation"""
    print("\nTesting content moderation...")
    
    # Create a temporary test video with a safe filename
    with tempfile.NamedTemporaryFile(suffix='_safe_video.mp4', delete=False) as tmp_file:
        test_video_path = tmp_file.name
    
    try:
        # Create test video
        create_test_video(test_video_path)
        print(f"Created test video: {test_video_path}")
        
        # Create mock metadata
        metadata = {
            "duration": 5.0,
            "width": 640,
            "height": 480,
            "fps": 30,
            "file_size": 1024 * 1024,  # 1MB
            "detailed": {
                "format": {
                    "tags": {
                        "title": "Safe Test Video",
                        "comment": "This is a test video for quality analysis"
                    }
                }
            }
        }
        
        # Analyze content safety
        safety_result = await content_moderator.analyze_content_safety(test_video_path, metadata)
        
        print("Content Moderation Results:")
        print(f"  Overall Safety Score: {safety_result.get('overall_score', 0):.3f}")
        print(f"  Is Safe: {safety_result.get('is_safe', False)}")
        print(f"  Confidence: {safety_result.get('confidence', 0):.3f}")
        print(f"  Flags Count: {len(safety_result.get('flags', []))}")
        print(f"  Moderation Action: {safety_result.get('moderation_action', 'unknown')}")
        print(f"  Analysis Methods: {safety_result.get('analysis_methods', [])}")
        
        # Check if we got reasonable results
        assert 'overall_score' in safety_result
        assert 'is_safe' in safety_result
        assert 'confidence' in safety_result
        assert 'flags' in safety_result
        assert 'moderation_action' in safety_result
        assert len(safety_result.get('analysis_methods', [])) > 0
        
        print("✓ Content moderation test passed!")
        return safety_result
        
    finally:
        # Clean up
        if os.path.exists(test_video_path):
            os.unlink(test_video_path)


async def test_weighted_scoring():
    """Test weighted quality scoring"""
    print("\nTesting weighted quality scoring...")
    
    # Mock quality metrics
    quality_metrics = {
        "sharpness_score": 0.8,
        "brightness_score": 0.7,
        "contrast_score": 0.9,
        "noise_score": 0.6,
        "overall_score": 0.75
    }
    
    # Mock safety metrics
    safety_metrics = {
        "overall_score": 0.9,
        "is_safe": True,
        "confidence": 0.8
    }
    
    # Mock metadata
    metadata = {
        "width": 1920,
        "height": 1080,
        "fps": 30,
        "duration": 120
    }
    
    # Calculate weighted score
    weighted_result = quality_scorer.calculate_weighted_score(
        quality_metrics, safety_metrics, metadata
    )
    
    print("Weighted Quality Scoring Results:")
    print(f"  Overall Score: {weighted_result.get('overall_score', 0):.3f}")
    print(f"  Technical Quality: {weighted_result.get('technical_quality', 0):.3f}")
    print(f"  Visual Appeal: {weighted_result.get('visual_appeal', 0):.3f}")
    print(f"  Content Quality: {weighted_result.get('content_quality', 0):.3f}")
    print(f"  Safety Contribution: {weighted_result.get('safety_contribution', 0):.3f}")
    print(f"  Quality Category: {weighted_result.get('quality_category', 'unknown')}")
    
    # Check if we got reasonable results
    assert 'overall_score' in weighted_result
    assert 'technical_quality' in weighted_result
    assert 'visual_appeal' in weighted_result
    assert 'content_quality' in weighted_result
    assert 'safety_contribution' in weighted_result
    assert 'quality_category' in weighted_result
    
    print("✓ Weighted quality scoring test passed!")
    return weighted_result


async def main():
    """Run all tests"""
    print("Starting Quality Analysis and Content Moderation Tests")
    print("=" * 60)
    
    try:
        # Test quality analysis
        quality_result = await test_quality_analysis()
        
        # Test content moderation
        safety_result = await test_content_moderation()
        
        # Test weighted scoring
        weighted_result = await test_weighted_scoring()
        
        print("\n" + "=" * 60)
        print("All tests passed successfully! ✓")
        print("\nSummary:")
        print(f"  Quality Analysis: Overall Score = {quality_result.get('overall_score', 0):.3f}")
        print(f"  Content Moderation: Safety Score = {safety_result.get('overall_score', 0):.3f}")
        print(f"  Weighted Scoring: Final Score = {weighted_result.get('overall_score', 0):.3f}")
        
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())