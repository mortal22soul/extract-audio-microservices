package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/video-converter/tests/integration/utils"
)

// VideoUploadResponse represents the response from video upload
type VideoUploadResponse struct {
	VideoID  string `json:"video_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

// VideoStatusResponse represents the response from video status check
type VideoStatusResponse struct {
	VideoID    string  `json:"video_id"`
	Status     string  `json:"status"`
	Progress   int     `json:"progress"`
	MP3FileID  string  `json:"mp3_file_id,omitempty"`
	Error      string  `json:"error,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// RealtimeMessage represents WebSocket messages
type RealtimeMessage struct {
	Type    string      `json:"type"`
	VideoID string      `json:"video_id"`
	Data    interface{} `json:"data"`
}

func TestCompleteVideoConversionWorkflow(t *testing.T) {
	config := utils.GetTestConfig()
	
	// Wait for services to be ready
	utils.WaitForServices(t, config)
	
	// Setup databases
	dbConns := utils.SetupDatabases(t, config)
	defer dbConns.CleanupDatabases(t)

	// Create test user
	testUser := utils.CreateTestUser(t, config)

	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Step 1: Upload video file
		videoID := uploadTestVideo(t, config, testUser)
		assert.NotEmpty(t, videoID)

		// Step 2: Verify video status is "uploaded"
		status := getVideoStatus(t, config, testUser, videoID)
		assert.Equal(t, "uploaded", status.Status)

		// Step 3: Connect to WebSocket for real-time updates
		wsConn := connectWebSocket(t, config, testUser)
		defer wsConn.Close()

		// Step 4: Wait for conversion to start and complete
		waitForConversionCompletion(t, wsConn, videoID, 60*time.Second)

		// Step 5: Verify final status is "completed"
		finalStatus := getVideoStatus(t, config, testUser, videoID)
		assert.Equal(t, "completed", finalStatus.Status)
		assert.Equal(t, 100, finalStatus.Progress)
		assert.NotEmpty(t, finalStatus.MP3FileID)

		// Step 6: Download converted MP3 file
		mp3Data := downloadMP3File(t, config, testUser, videoID)
		assert.NotEmpty(t, mp3Data)
		assert.Greater(t, len(mp3Data), 0)

		// Step 7: Verify video appears in user's history
		history := getUserVideoHistory(t, config, testUser)
		assert.NotEmpty(t, history)
		
		found := false
		for _, video := range history {
			if video["video_id"] == videoID {
				found = true
				assert.Equal(t, "completed", video["status"])
				break
			}
		}
		assert.True(t, found, "Video should appear in user's history")
	})

	t.Run("MultipleVideosWorkflow", func(t *testing.T) {
		// Upload multiple videos concurrently
		videoCount := 3
		videoIDs := make([]string, videoCount)
		
		for i := 0; i < videoCount; i++ {
			videoIDs[i] = uploadTestVideo(t, config, testUser)
		}

		// Connect to WebSocket
		wsConn := connectWebSocket(t, config, testUser)
		defer wsConn.Close()

		// Wait for all conversions to complete
		completedVideos := make(map[string]bool)
		timeout := time.After(120 * time.Second)
		
		for len(completedVideos) < videoCount {
			select {
			case <-timeout:
				t.Fatalf("Timeout waiting for video conversions to complete")
			default:
				// Check WebSocket messages
				wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
				var msg RealtimeMessage
				err := wsConn.ReadJSON(&msg)
				if err != nil {
					continue
				}

				if msg.Type == "conversion_complete" {
					completedVideos[msg.VideoID] = true
				}
			}
		}

		// Verify all videos are completed
		for _, videoID := range videoIDs {
			status := getVideoStatus(t, config, testUser, videoID)
			assert.Equal(t, "completed", status.Status)
		}
	})

	t.Run("ErrorHandlingWorkflow", func(t *testing.T) {
		// Upload invalid file (should fail conversion)
		invalidVideoID := uploadInvalidVideo(t, config, testUser)
		
		// Connect to WebSocket
		wsConn := connectWebSocket(t, config, testUser)
		defer wsConn.Close()

		// Wait for error notification
		timeout := time.After(30 * time.Second)
		errorReceived := false
		
		for !errorReceived {
			select {
			case <-timeout:
				t.Fatalf("Timeout waiting for error notification")
			default:
				wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
				var msg RealtimeMessage
				err := wsConn.ReadJSON(&msg)
				if err != nil {
					continue
				}

				if msg.Type == "conversion_error" && msg.VideoID == invalidVideoID {
					errorReceived = true
				}
			}
		}

		// Verify video status is "failed"
		status := getVideoStatus(t, config, testUser, invalidVideoID)
		assert.Equal(t, "failed", status.Status)
		assert.NotEmpty(t, status.Error)
	})
}

func uploadTestVideo(t *testing.T, config *utils.TestConfig, user *utils.TestUser) string {
	// Create test video content
	videoContent := utils.CreateTestVideoFile(t)
	
	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	part, err := writer.CreateFormFile("video", "test_video.mp4")
	require.NoError(t, err)
	
	_, err = part.Write(videoContent)
	require.NoError(t, err)
	
	err = writer.Close()
	require.NoError(t, err)

	// Create HTTP request
	req, err := http.NewRequest("POST", config.GatewayURL+"/api/v1/videos/upload", &buf)
	require.NoError(t, err)
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+user.Token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var uploadResp VideoUploadResponse
	err = json.NewDecoder(resp.Body).Decode(&uploadResp)
	require.NoError(t, err)

	return uploadResp.VideoID
}

func uploadInvalidVideo(t *testing.T, config *utils.TestConfig, user *utils.TestUser) string {
	// Create invalid video content (just text)
	invalidContent := []byte("This is not a video file")
	
	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	part, err := writer.CreateFormFile("video", "invalid.txt")
	require.NoError(t, err)
	
	_, err = part.Write(invalidContent)
	require.NoError(t, err)
	
	err = writer.Close()
	require.NoError(t, err)

	// Create HTTP request
	req, err := http.NewRequest("POST", config.GatewayURL+"/api/v1/videos/upload", &buf)
	require.NoError(t, err)
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+user.Token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should still accept upload, but conversion will fail
	var uploadResp VideoUploadResponse
	err = json.NewDecoder(resp.Body).Decode(&uploadResp)
	require.NoError(t, err)

	return uploadResp.VideoID
}

func getVideoStatus(t *testing.T, config *utils.TestConfig, user *utils.TestUser, videoID string) *VideoStatusResponse {
	req, err := http.NewRequest("GET", config.GatewayURL+"/api/v1/videos/"+videoID+"/status", nil)
	require.NoError(t, err)
	
	req.Header.Set("Authorization", "Bearer "+user.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var statusResp VideoStatusResponse
	err = json.NewDecoder(resp.Body).Decode(&statusResp)
	require.NoError(t, err)

	return &statusResp
}

func connectWebSocket(t *testing.T, config *utils.TestConfig, user *utils.TestUser) *websocket.Conn {
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws://localhost:3001/socket.io/?EIO=4&transport=websocket&token=" + user.Token
	
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	
	return conn
}

func waitForConversionCompletion(t *testing.T, wsConn *websocket.Conn, videoID string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
		
		var msg RealtimeMessage
		err := wsConn.ReadJSON(&msg)
		if err != nil {
			continue
		}

		if msg.VideoID == videoID {
			switch msg.Type {
			case "conversion_progress":
				t.Logf("Conversion progress for %s: %v", videoID, msg.Data)
			case "conversion_complete":
				t.Logf("Conversion completed for %s", videoID)
				return
			case "conversion_error":
				t.Fatalf("Conversion failed for %s: %v", videoID, msg.Data)
			}
		}
	}
	
	t.Fatalf("Timeout waiting for conversion completion of video %s", videoID)
}

func downloadMP3File(t *testing.T, config *utils.TestConfig, user *utils.TestUser, videoID string) []byte {
	req, err := http.NewRequest("GET", config.GatewayURL+"/api/v1/videos/"+videoID+"/download", nil)
	require.NoError(t, err)
	
	req.Header.Set("Authorization", "Bearer "+user.Token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "audio/mpeg", resp.Header.Get("Content-Type"))

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return data
}

func getUserVideoHistory(t *testing.T, config *utils.TestConfig, user *utils.TestUser) []map[string]interface{} {
	req, err := http.NewRequest("GET", config.GatewayURL+"/api/v1/videos/history", nil)
	require.NoError(t, err)
	
	req.Header.Set("Authorization", "Bearer "+user.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var history []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&history)
	require.NoError(t, err)

	return history
}