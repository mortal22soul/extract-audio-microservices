package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	analyticspb "github.com/video-converter/shared/proto/gen/go/shared/proto"
	commonpb "github.com/video-converter/shared/proto/gen/go/shared/proto"
	"github.com/video-converter/tests/integration/utils"
)

func TestAnalyticsServiceGRPC(t *testing.T) {
	config := utils.GetTestConfig()
	
	// Wait for services to be ready
	utils.WaitForServices(t, config)
	
	// Setup gRPC connections
	grpcConns := utils.SetupGRPCConnections(t, config)
	defer grpcConns.CleanupGRPCConnections()
	
	// Setup databases
	dbConns := utils.SetupDatabases(t, config)
	defer dbConns.CleanupDatabases(t)

	// Create analytics service client
	analyticsClient := analyticspb.NewAnalyticsServiceClient(grpcConns.AnalyticsConn)

	t.Run("AnalyzeVideo", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req := &analyticspb.VideoAnalysisRequest{
			VideoId: "test_video_123",
			UserId:  "test_user_456",
			VideoPath: "/tmp/test_video.mp4",
			Metadata: &commonpb.VideoMetadata{
				Filename:        "test_video.mp4",
				SizeBytes:       1024000,
				DurationSeconds: 120,
				Resolution:      "1920x1080",
				Codec:          "h264",
				Bitrate:        2000,
				Format:         commonpb.VideoFormat_VIDEO_FORMAT_MP4,
			},
		}

		resp, err := analyticsClient.AnalyzeVideo(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test_video_123", resp.VideoId)
		assert.NotNil(t, resp.Quality)
		assert.NotNil(t, resp.Safety)
		
		// Check quality metrics
		assert.GreaterOrEqual(t, resp.Quality.OverallScore, float32(0.0))
		assert.LessOrEqual(t, resp.Quality.OverallScore, float32(1.0))
		assert.NotEmpty(t, resp.Quality.ResolutionCategory)
		
		// Check safety score
		assert.GreaterOrEqual(t, resp.Safety.OverallScore, float32(0.0))
		assert.LessOrEqual(t, resp.Safety.OverallScore, float32(1.0))
	})

	t.Run("GetQualityMetrics", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		req := &analyticspb.QualityRequest{
			VideoId:   "test_video_quality",
			VideoPath: "/tmp/test_video.mp4",
		}

		resp, err := analyticsClient.GetQualityMetrics(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp.Quality)
		
		quality := resp.Quality
		assert.GreaterOrEqual(t, quality.SharpnessScore, float32(0.0))
		assert.LessOrEqual(t, quality.SharpnessScore, float32(1.0))
		assert.GreaterOrEqual(t, quality.BrightnessScore, float32(0.0))
		assert.LessOrEqual(t, quality.BrightnessScore, float32(1.0))
		assert.GreaterOrEqual(t, quality.ContrastScore, float32(0.0))
		assert.LessOrEqual(t, quality.ContrastScore, float32(1.0))
		assert.GreaterOrEqual(t, quality.OverallScore, float32(0.0))
		assert.LessOrEqual(t, quality.OverallScore, float32(1.0))
		assert.Contains(t, []string{"low", "medium", "high", "ultra"}, quality.ResolutionCategory)
	})

	t.Run("CheckContentSafety", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		req := &analyticspb.SafetyRequest{
			VideoId:   "test_video_safety",
			VideoPath: "/tmp/test_video.mp4",
		}

		resp, err := analyticsClient.CheckContentSafety(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp.Safety)
		
		safety := resp.Safety
		assert.GreaterOrEqual(t, safety.OverallScore, float32(0.0))
		assert.LessOrEqual(t, safety.OverallScore, float32(1.0))
		
		// Check safety flags if any
		for _, flag := range safety.Flags {
			assert.NotEmpty(t, flag.Category)
			assert.GreaterOrEqual(t, flag.Confidence, float32(0.0))
			assert.LessOrEqual(t, flag.Confidence, float32(1.0))
		}
	})

	t.Run("GenerateThumbnails", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req := &analyticspb.ThumbnailRequest{
			VideoId:   "test_video_thumbnails",
			VideoPath: "/tmp/test_video.mp4",
			Count:     3,
		}

		resp, err := analyticsClient.GenerateThumbnails(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Thumbnails)
		assert.LessOrEqual(t, len(resp.Thumbnails), 3)
		
		for _, thumbnail := range resp.Thumbnails {
			assert.NotEmpty(t, thumbnail.Url)
			assert.GreaterOrEqual(t, thumbnail.TimestampSeconds, int32(0))
			assert.Greater(t, thumbnail.Width, int32(0))
			assert.Greater(t, thumbnail.Height, int32(0))
		}
	})

	t.Run("GenerateThumbnailsWithTimestamps", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		timestamps := []int32{10, 30, 60}
		req := &analyticspb.ThumbnailRequest{
			VideoId:           "test_video_timestamps",
			VideoPath:         "/tmp/test_video.mp4",
			TimestampSeconds:  timestamps,
		}

		resp, err := analyticsClient.GenerateThumbnails(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, len(timestamps), len(resp.Thumbnails))
		
		for i, thumbnail := range resp.Thumbnails {
			assert.Equal(t, timestamps[i], thumbnail.TimestampSeconds)
			assert.NotEmpty(t, thumbnail.Url)
		}
	})

	t.Run("GetRecommendations", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		req := &analyticspb.RecommendationRequest{
			UserId:           "test_user_recommendations",
			Limit:            5,
			ExcludeVideoIds:  []string{"exclude_video_1", "exclude_video_2"},
		}

		resp, err := analyticsClient.GetRecommendations(ctx, req)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(resp.Recommendations), 5)
		
		// Verify excluded videos are not in recommendations
		for _, rec := range resp.Recommendations {
			assert.NotContains(t, req.ExcludeVideoIds, rec.VideoId)
			assert.NotEmpty(t, rec.VideoId)
			assert.GreaterOrEqual(t, rec.SimilarityScore, float32(0.0))
			assert.LessOrEqual(t, rec.SimilarityScore, float32(1.0))
		}
	})

	t.Run("GetRecommendationsEmptyResult", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		req := &analyticspb.RecommendationRequest{
			UserId: "nonexistent_user",
			Limit:  10,
		}

		resp, err := analyticsClient.GetRecommendations(ctx, req)
		require.NoError(t, err)
		// Should return empty recommendations for non-existent user
		assert.Empty(t, resp.Recommendations)
	})

	t.Run("InvalidVideoPath", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := &analyticspb.QualityRequest{
			VideoId:   "invalid_video",
			VideoPath: "/nonexistent/path/video.mp4",
		}

		resp, err := analyticsClient.GetQualityMetrics(ctx, req)
		// Should handle invalid path gracefully
		if err == nil {
			assert.NotNil(t, resp.Error)
			assert.NotEmpty(t, resp.Error.Message)
		}
	})
}