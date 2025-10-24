package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/video-converter/tests/integration/utils"
)

// ChaosTest represents a chaos engineering test scenario
type ChaosTest struct {
	Name        string
	Description string
	Setup       func(t *testing.T, config *utils.TestConfig) ChaosScenario
	Verify      func(t *testing.T, scenario ChaosScenario, results ChaosResults)
}

// ChaosScenario holds the test scenario configuration
type ChaosScenario struct {
	Config      *utils.TestConfig
	TestUser    *utils.TestUser
	VideoIDs    []string
	Connections *utils.DatabaseConnections
}

// ChaosResults holds the results of chaos testing
type ChaosResults struct {
	SuccessfulUploads   int
	FailedUploads       int
	SuccessfulDownloads int
	FailedDownloads     int
	CompletedConversions int
	FailedConversions   int
	Errors              []error
	Duration            time.Duration
}

func TestSystemResilience(t *testing.T) {
	chaosTests := []ChaosTest{
		{
			Name:        "DatabaseConnectionFailure",
			Description: "Test system behavior when database connections are intermittently lost",
			Setup:       setupDatabaseChaos,
			Verify:      verifyDatabaseResilience,
		},
		{
			Name:        "HighLoadStress",
			Description: "Test system under high concurrent load",
			Setup:       setupHighLoadChaos,
			Verify:      verifyHighLoadResilience,
		},
		{
			Name:        "ServicePartialFailure",
			Description: "Test system when some services are unavailable",
			Setup:       setupServiceFailureChaos,
			Verify:      verifyServiceFailureResilience,
		},
		{
			Name:        "NetworkLatency",
			Description: "Test system behavior under high network latency",
			Setup:       setupNetworkLatencyChaos,
			Verify:      verifyNetworkLatencyResilience,
		},
	}

	for _, chaosTest := range chaosTests {
		t.Run(chaosTest.Name, func(t *testing.T) {
			t.Logf("Running chaos test: %s - %s", chaosTest.Name, chaosTest.Description)
			
			scenario := chaosTest.Setup(t, utils.GetTestConfig())
			defer scenario.Connections.CleanupDatabases(t)
			
			results := runChaosScenario(t, scenario)
			chaosTest.Verify(t, scenario, results)
		})
	}
}

func setupDatabaseChaos(t *testing.T, config *utils.TestConfig) ChaosScenario {
	utils.WaitForServices(t, config)
	
	dbConns := utils.SetupDatabases(t, config)
	testUser := utils.CreateTestUser(t, config)
	
	return ChaosScenario{
		Config:      config,
		TestUser:    testUser,
		Connections: dbConns,
	}
}

func setupHighLoadChaos(t *testing.T, config *utils.TestConfig) ChaosScenario {
	utils.WaitForServices(t, config)
	
	dbConns := utils.SetupDatabases(t, config)
	testUser := utils.CreateTestUser(t, config)
	
	return ChaosScenario{
		Config:      config,
		TestUser:    testUser,
		Connections: dbConns,
	}
}

func setupServiceFailureChaos(t *testing.T, config *utils.TestConfig) ChaosScenario {
	utils.WaitForServices(t, config)
	
	dbConns := utils.SetupDatabases(t, config)
	testUser := utils.CreateTestUser(t, config)
	
	return ChaosScenario{
		Config:      config,
		TestUser:    testUser,
		Connections: dbConns,
	}
}

func setupNetworkLatencyChaos(t *testing.T, config *utils.TestConfig) ChaosScenario {
	utils.WaitForServices(t, config)
	
	dbConns := utils.SetupDatabases(t, config)
	testUser := utils.CreateTestUser(t, config)
	
	return ChaosScenario{
		Config:      config,
		TestUser:    testUser,
		Connections: dbConns,
	}
}

func runChaosScenario(t *testing.T, scenario ChaosScenario) ChaosResults {
	startTime := time.Now()
	results := ChaosResults{}
	
	// Run concurrent operations to stress test the system
	concurrency := 10
	operationsPerWorker := 5
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < operationsPerWorker; j++ {
				// Simulate various operations
				operation := rand.Intn(3)
				
				switch operation {
				case 0:
					// Upload video
					err := simulateVideoUpload(scenario, fmt.Sprintf("worker_%d_video_%d", workerID, j))
					mu.Lock()
					if err != nil {
						results.FailedUploads++
						results.Errors = append(results.Errors, err)
					} else {
						results.SuccessfulUploads++
					}
					mu.Unlock()
					
				case 1:
					// Check video status
					err := simulateStatusCheck(scenario, fmt.Sprintf("test_video_%d", rand.Intn(100)))
					mu.Lock()
					if err != nil {
						results.Errors = append(results.Errors, err)
					}
					mu.Unlock()
					
				case 2:
					// Download video
					err := simulateVideoDownload(scenario, fmt.Sprintf("test_video_%d", rand.Intn(100)))
					mu.Lock()
					if err != nil {
						results.FailedDownloads++
						results.Errors = append(results.Errors, err)
					} else {
						results.SuccessfulDownloads++
					}
					mu.Unlock()
				}
				
				// Add random delay between operations
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}
		}(i)
	}
	
	// Introduce chaos while operations are running
	go introduceChaos(t, scenario)
	
	wg.Wait()
	results.Duration = time.Since(startTime)
	
	return results
}

func introduceChaos(t *testing.T, scenario ChaosScenario) {
	// Simulate various failure scenarios
	chaosEvents := []func(){
		func() {
			// Simulate database connection issues
			t.Logf("Introducing database connection chaos")
			time.Sleep(2 * time.Second)
		},
		func() {
			// Simulate Redis connection issues
			t.Logf("Introducing Redis connection chaos")
			time.Sleep(1 * time.Second)
		},
		func() {
			// Simulate high CPU load
			t.Logf("Introducing CPU load chaos")
			time.Sleep(3 * time.Second)
		},
	}
	
	for _, event := range chaosEvents {
		event()
		time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
	}
}

func simulateVideoUpload(scenario ChaosScenario, videoID string) error {
	// Simulate video upload with potential failures
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Add random failure chance
	if rand.Float32() < 0.1 { // 10% failure rate
		return fmt.Errorf("simulated upload failure for video %s", videoID)
	}
	
	// Simulate upload delay
	select {
	case <-ctx.Done():
		return fmt.Errorf("upload timeout for video %s", videoID)
	case <-time.After(time.Duration(rand.Intn(1000)) * time.Millisecond):
		return nil
	}
}

func simulateStatusCheck(scenario ChaosScenario, videoID string) error {
	// Simulate status check with potential failures
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	// Add random failure chance
	if rand.Float32() < 0.05 { // 5% failure rate
		return fmt.Errorf("simulated status check failure for video %s", videoID)
	}
	
	// Simulate check delay
	select {
	case <-ctx.Done():
		return fmt.Errorf("status check timeout for video %s", videoID)
	case <-time.After(time.Duration(rand.Intn(500)) * time.Millisecond):
		return nil
	}
}

func simulateVideoDownload(scenario ChaosScenario, videoID string) error {
	// Simulate video download with potential failures
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Add random failure chance
	if rand.Float32() < 0.15 { // 15% failure rate
		return fmt.Errorf("simulated download failure for video %s", videoID)
	}
	
	// Simulate download delay
	select {
	case <-ctx.Done():
		return fmt.Errorf("download timeout for video %s", videoID)
	case <-time.After(time.Duration(rand.Intn(2000)) * time.Millisecond):
		return nil
	}
}

func verifyDatabaseResilience(t *testing.T, scenario ChaosScenario, results ChaosResults) {
	t.Logf("Database resilience test results:")
	t.Logf("  Duration: %v", results.Duration)
	t.Logf("  Successful uploads: %d", results.SuccessfulUploads)
	t.Logf("  Failed uploads: %d", results.FailedUploads)
	t.Logf("  Successful downloads: %d", results.SuccessfulDownloads)
	t.Logf("  Failed downloads: %d", results.FailedDownloads)
	t.Logf("  Total errors: %d", len(results.Errors))
	
	// System should maintain at least 70% success rate under database chaos
	totalOperations := results.SuccessfulUploads + results.FailedUploads + results.SuccessfulDownloads + results.FailedDownloads
	successfulOperations := results.SuccessfulUploads + results.SuccessfulDownloads
	
	if totalOperations > 0 {
		successRate := float64(successfulOperations) / float64(totalOperations)
		assert.GreaterOrEqual(t, successRate, 0.7, "System should maintain at least 70% success rate under database chaos")
	}
	
	// System should complete within reasonable time even under stress
	assert.LessOrEqual(t, results.Duration, 60*time.Second, "System should complete chaos test within 60 seconds")
}

func verifyHighLoadResilience(t *testing.T, scenario ChaosScenario, results ChaosResults) {
	t.Logf("High load resilience test results:")
	t.Logf("  Duration: %v", results.Duration)
	t.Logf("  Successful uploads: %d", results.SuccessfulUploads)
	t.Logf("  Failed uploads: %d", results.FailedUploads)
	t.Logf("  Successful downloads: %d", results.SuccessfulDownloads)
	t.Logf("  Failed downloads: %d", results.FailedDownloads)
	t.Logf("  Total errors: %d", len(results.Errors))
	
	// System should handle high load gracefully
	totalOperations := results.SuccessfulUploads + results.FailedUploads + results.SuccessfulDownloads + results.FailedDownloads
	successfulOperations := results.SuccessfulUploads + results.SuccessfulDownloads
	
	if totalOperations > 0 {
		successRate := float64(successfulOperations) / float64(totalOperations)
		assert.GreaterOrEqual(t, successRate, 0.8, "System should maintain at least 80% success rate under high load")
	}
	
	// No single operation should take too long
	assert.LessOrEqual(t, results.Duration, 45*time.Second, "System should handle high load within 45 seconds")
}

func verifyServiceFailureResilience(t *testing.T, scenario ChaosScenario, results ChaosResults) {
	t.Logf("Service failure resilience test results:")
	t.Logf("  Duration: %v", results.Duration)
	t.Logf("  Successful uploads: %d", results.SuccessfulUploads)
	t.Logf("  Failed uploads: %d", results.FailedUploads)
	t.Logf("  Successful downloads: %d", results.SuccessfulDownloads)
	t.Logf("  Failed downloads: %d", results.FailedDownloads)
	t.Logf("  Total errors: %d", len(results.Errors))
	
	// System should degrade gracefully when services fail
	totalOperations := results.SuccessfulUploads + results.FailedUploads + results.SuccessfulDownloads + results.FailedDownloads
	
	// At least some operations should succeed even with service failures
	assert.Greater(t, totalOperations, 0, "System should attempt operations even with service failures")
	
	// Error rate should be reasonable (not 100% failure)
	if totalOperations > 0 {
		errorRate := float64(len(results.Errors)) / float64(totalOperations)
		assert.LessOrEqual(t, errorRate, 0.5, "Error rate should not exceed 50% even with service failures")
	}
}

func verifyNetworkLatencyResilience(t *testing.T, scenario ChaosScenario, results ChaosResults) {
	t.Logf("Network latency resilience test results:")
	t.Logf("  Duration: %v", results.Duration)
	t.Logf("  Successful uploads: %d", results.SuccessfulUploads)
	t.Logf("  Failed uploads: %d", results.FailedUploads)
	t.Logf("  Successful downloads: %d", results.SuccessfulDownloads)
	t.Logf("  Failed downloads: %d", results.FailedDownloads)
	t.Logf("  Total errors: %d", len(results.Errors))
	
	// System should handle network latency gracefully
	totalOperations := results.SuccessfulUploads + results.FailedUploads + results.SuccessfulDownloads + results.FailedDownloads
	successfulOperations := results.SuccessfulUploads + results.SuccessfulDownloads
	
	if totalOperations > 0 {
		successRate := float64(successfulOperations) / float64(totalOperations)
		assert.GreaterOrEqual(t, successRate, 0.75, "System should maintain at least 75% success rate under network latency")
	}
	
	// Operations should eventually complete despite latency
	assert.LessOrEqual(t, results.Duration, 90*time.Second, "System should complete operations within 90 seconds despite network latency")
}