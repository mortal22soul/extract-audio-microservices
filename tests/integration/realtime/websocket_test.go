package realtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/video-converter/tests/integration/utils"
)

// WebSocketMessage represents messages sent/received via WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	VideoID string      `json:"video_id,omitempty"`
	UserID  string      `json:"user_id,omitempty"`
	Data    interface{} `json:"data"`
}

// ProgressData represents conversion progress data
type ProgressData struct {
	Progress    int    `json:"progress"`
	Status      string `json:"status"`
	EstimatedTime int  `json:"estimated_time,omitempty"`
}

func TestWebSocketRealTimeNotifications(t *testing.T) {
	config := utils.GetTestConfig()
	
	// Wait for services to be ready
	utils.WaitForServices(t, config)
	
	// Setup databases
	dbConns := utils.SetupDatabases(t, config)
	defer dbConns.CleanupDatabases(t)

	// Create test user
	testUser := utils.CreateTestUser(t, config)

	t.Run("WebSocketConnection", func(t *testing.T) {
		// Connect to WebSocket
		wsConn := connectWebSocket(t, config, testUser)
		defer wsConn.Close()

		// Send ping and expect pong
		err := wsConn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
		require.NoError(t, err)

		// Read response
		wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, message, err := wsConn.ReadMessage()
		require.NoError(t, err)

		var response WebSocketMessage
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)
		assert.Equal(t, "pong", response.Type)
	})

	t.Run("ConversionProgressNotifications", func(t *testing.T) {
		// Connect to WebSocket
		wsConn := connectWebSocket(t, config, testUser)
		defer wsConn.Close()

		// Simulate conversion progress via Redis pub/sub
		videoID := "test_video_progress_123"
		
		// Publish progress updates
		go func() {
			time.Sleep(1 * time.Second) // Wait for WebSocket to be ready
			
			progressUpdates := []int{25, 50, 75, 100}
			for _, progress := range progressUpdates {
				publishProgressUpdate(t, dbConns.Redis, testUser.ID, videoID, progress)
				time.Sleep(500 * time.Millisecond)
			}
		}()

		// Listen for progress notifications
		receivedProgress := make([]int, 0)
		timeout := time.After(10 * time.Second)
		
		for len(receivedProgress) < 4 {
			select {
			case <-timeout:
				t.Fatalf("Timeout waiting for progress notifications. Received: %v", receivedProgress)
			default:
				wsConn.SetReadDeadline(time.Now().Add(2 * time.Second))
				_, message, err := wsConn.ReadMessage()
				if err != nil {
					continue
				}

				var wsMsg WebSocketMessage
				err = json.Unmarshal(message, &wsMsg)
				if err != nil {
					continue
				}

				if wsMsg.Type == "conversion_progress" && wsMsg.VideoID == videoID {
					progressData, ok := wsMsg.Data.(map[string]interface{})
					if ok {
						if progress, exists := progressData["progress"]; exists {
							if progressFloat, ok := progress.(float64); ok {
								receivedProgress = append(receivedProgress, int(progressFloat))
							}
						}
					}
				}
			}
		}

		// Verify all progress updates were received
		assert.Equal(t, []int{25, 50, 75, 100}, receivedProgress)
	})

	t.Run("ConversionCompleteNotification", func(t *testing.T) {
		// Connect to WebSocket
		wsConn := connectWebSocket(t, config, testUser)
		defer wsConn.Close()

		videoID := "test_video_complete_456"
		
		// Publish completion notification
		go func() {
			time.Sleep(1 * time.Second)
			publishCompletionNotification(t, dbConns.Redis, testUser.ID, videoID)
		}()

		// Wait for completion notification
		timeout := time.After(5 * time.Second)
		completed := false
		
		for !completed {
			select {
			case <-timeout:
				t.Fatalf("Timeout waiting for completion notification")
			default:
				wsConn.SetReadDeadline(time.Now().Add(2 * time.Second))
				_, message, err := wsConn.ReadMessage()
				if err != nil {
					continue
				}

				var wsMsg WebSocketMessage
				err = json.Unmarshal(message, &wsMsg)
				if err != nil {
					continue
				}

				if wsMsg.Type == "conversion_complete" && wsMsg.VideoID == videoID {
					completed = true
					assert.Equal(t, testUser.ID, wsMsg.UserID)
				}
			}
		}
	})

	t.Run("ConversionErrorNotification", func(t *testing.T) {
		// Connect to WebSocket
		wsConn := connectWebSocket(t, config, testUser)
		defer wsConn.Close()

		videoID := "test_video_error_789"
		errorMessage := "FFmpeg conversion failed: invalid codec"
		
		// Publish error notification
		go func() {
			time.Sleep(1 * time.Second)
			publishErrorNotification(t, dbConns.Redis, testUser.ID, videoID, errorMessage)
		}()

		// Wait for error notification
		timeout := time.After(5 * time.Second)
		errorReceived := false
		
		for !errorReceived {
			select {
			case <-timeout:
				t.Fatalf("Timeout waiting for error notification")
			default:
				wsConn.SetReadDeadline(time.Now().Add(2 * time.Second))
				_, message, err := wsConn.ReadMessage()
				if err != nil {
					continue
				}

				var wsMsg WebSocketMessage
				err = json.Unmarshal(message, &wsMsg)
				if err != nil {
					continue
				}

				if wsMsg.Type == "conversion_error" && wsMsg.VideoID == videoID {
					errorReceived = true
					assert.Equal(t, testUser.ID, wsMsg.UserID)
					
					errorData, ok := wsMsg.Data.(map[string]interface{})
					if ok {
						assert.Equal(t, errorMessage, errorData["error"])
					}
				}
			}
		}
	})

	t.Run("MultipleClientsNotifications", func(t *testing.T) {
		// Connect multiple WebSocket clients for the same user
		wsConn1 := connectWebSocket(t, config, testUser)
		defer wsConn1.Close()
		
		wsConn2 := connectWebSocket(t, config, testUser)
		defer wsConn2.Close()

		videoID := "test_video_multi_clients"
		
		// Publish notification
		go func() {
			time.Sleep(1 * time.Second)
			publishProgressUpdate(t, dbConns.Redis, testUser.ID, videoID, 50)
		}()

		// Both clients should receive the notification
		clients := []*websocket.Conn{wsConn1, wsConn2}
		receivedCount := 0
		timeout := time.After(5 * time.Second)
		
		for receivedCount < 2 {
			select {
			case <-timeout:
				t.Fatalf("Timeout waiting for notifications on multiple clients")
			default:
				for _, conn := range clients {
					conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
					_, message, err := conn.ReadMessage()
					if err != nil {
						continue
					}

					var wsMsg WebSocketMessage
					err = json.Unmarshal(message, &wsMsg)
					if err != nil {
						continue
					}

					if wsMsg.Type == "conversion_progress" && wsMsg.VideoID == videoID {
						receivedCount++
					}
				}
			}
		}

		assert.Equal(t, 2, receivedCount)
	})

	t.Run("UserIsolation", func(t *testing.T) {
		// Create second test user
		testUser2 := utils.CreateTestUser(t, config)
		
		// Connect WebSocket for both users
		wsConn1 := connectWebSocket(t, config, testUser)
		defer wsConn1.Close()
		
		wsConn2 := connectWebSocket(t, config, testUser2)
		defer wsConn2.Close()

		videoID := "test_video_isolation"
		
		// Publish notification for user 1 only
		go func() {
			time.Sleep(1 * time.Second)
			publishProgressUpdate(t, dbConns.Redis, testUser.ID, videoID, 75)
		}()

		// Only user 1 should receive the notification
		user1Received := false
		user2Received := false
		timeout := time.After(3 * time.Second)
		
		for time.Now().Before(timeout.C) {
			// Check user 1 connection
			wsConn1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			_, message, err := wsConn1.ReadMessage()
			if err == nil {
				var wsMsg WebSocketMessage
				if json.Unmarshal(message, &wsMsg) == nil {
					if wsMsg.Type == "conversion_progress" && wsMsg.VideoID == videoID {
						user1Received = true
					}
				}
			}

			// Check user 2 connection
			wsConn2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			_, message, err = wsConn2.ReadMessage()
			if err == nil {
				var wsMsg WebSocketMessage
				if json.Unmarshal(message, &wsMsg) == nil {
					if wsMsg.Type == "conversion_progress" && wsMsg.VideoID == videoID {
						user2Received = true
					}
				}
			}
		}

		assert.True(t, user1Received, "User 1 should receive the notification")
		assert.False(t, user2Received, "User 2 should not receive the notification")
	})
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

func publishProgressUpdate(t *testing.T, redisClient *redis.Client, userID, videoID string, progress int) {
	ctx := context.Background()
	
	message := map[string]interface{}{
		"user_id":  userID,
		"video_id": videoID,
		"progress": progress,
		"status":   "processing",
	}
	
	messageJSON, err := json.Marshal(message)
	require.NoError(t, err)
	
	err = redisClient.Publish(ctx, "conversion:progress", messageJSON).Err()
	require.NoError(t, err)
}

func publishCompletionNotification(t *testing.T, redisClient *redis.Client, userID, videoID string) {
	ctx := context.Background()
	
	message := map[string]interface{}{
		"user_id":     userID,
		"video_id":    videoID,
		"status":      "completed",
		"mp3_file_id": "converted_file_123",
	}
	
	messageJSON, err := json.Marshal(message)
	require.NoError(t, err)
	
	err = redisClient.Publish(ctx, "conversion:complete", messageJSON).Err()
	require.NoError(t, err)
}

func publishErrorNotification(t *testing.T, redisClient *redis.Client, userID, videoID, errorMsg string) {
	ctx := context.Background()
	
	message := map[string]interface{}{
		"user_id":  userID,
		"video_id": videoID,
		"status":   "failed",
		"error":    errorMsg,
	}
	
	messageJSON, err := json.Marshal(message)
	require.NoError(t, err)
	
	err = redisClient.Publish(ctx, "conversion:error", messageJSON).Err()
	require.NoError(t, err)
}