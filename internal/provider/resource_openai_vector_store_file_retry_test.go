package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// mockReadFunc is a function type that simulates the read operation
type mockReadFunc func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics

// TestRetryLogicBehavior tests the actual retry behavior by simulating different scenarios
func TestRetryLogicBehavior(t *testing.T) {
	t.Run("SuccessOnFirstAttempt", func(t *testing.T) {
		// Track call count
		callCount := 0
		startTime := time.Now()

		// Create a mock read function that succeeds immediately
		mockRead := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			callCount++
			return nil // Success on first try
		}

		// Simulate retry logic
		result := simulateRetryLogic(mockRead, 5)

		elapsed := time.Since(startTime)

		// Assertions
		if result.HasError() {
			t.Errorf("Expected no error, got: %v", result)
		}
		if callCount != 1 {
			t.Errorf("Expected 1 call, got %d", callCount)
		}
		// Should complete almost immediately (no retries)
		if elapsed > 500*time.Millisecond {
			t.Errorf("Expected to complete quickly, took %v", elapsed)
		}
	})

	t.Run("SuccessAfterThreeRetries", func(t *testing.T) {
		callCount := 0
		startTime := time.Now()

		// Create a mock read function that fails 3 times then succeeds
		mockRead := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			callCount++
			if callCount < 3 {
				return diag.Errorf("Error reading vector store file: API error: No file found with id 'file-123' in vector store 'vs-456'")
			}
			return nil // Success on 3rd attempt
		}

		result := simulateRetryLogic(mockRead, 5)

		elapsed := time.Since(startTime)

		// Assertions
		if result.HasError() {
			t.Errorf("Expected no error after retries, got: %v", result)
		}
		if callCount != 3 {
			t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount)
		}
		// Should have waited: 1s + 2s = 3s (approximately)
		// Using a range to account for test execution time
		if elapsed < 2*time.Second || elapsed > 4*time.Second {
			t.Errorf("Expected ~3s elapsed (1s + 2s backoff), got %v", elapsed)
		}
	})

	t.Run("MaxRetriesExhausted", func(t *testing.T) {
		callCount := 0
		maxRetries := 5

		// Create a mock read function that always fails with retriable error
		mockRead := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			callCount++
			return diag.Errorf("Error reading vector store file: API error: No file found with id 'file-123' in vector store 'vs-456'")
		}

		result := simulateRetryLogic(mockRead, maxRetries)

		// Assertions
		if !result.HasError() {
			t.Error("Expected error after max retries exhausted")
		}
		if callCount != maxRetries {
			t.Errorf("Expected %d calls (max retries), got %d", maxRetries, callCount)
		}
	})

	t.Run("NonRetriableErrorFailsImmediately", func(t *testing.T) {
		callCount := 0
		startTime := time.Now()

		// Create a mock read function that fails with non-retriable error
		mockRead := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			callCount++
			return diag.Errorf("Error reading vector store file: API error: Unauthorized")
		}

		result := simulateRetryLogic(mockRead, 5)

		elapsed := time.Since(startTime)

		// Assertions
		if !result.HasError() {
			t.Error("Expected error for unauthorized")
		}
		if callCount != 1 {
			t.Errorf("Expected 1 call (no retries for non-retriable error), got %d", callCount)
		}
		// Should fail immediately without waiting
		if elapsed > 500*time.Millisecond {
			t.Errorf("Expected to fail immediately, took %v", elapsed)
		}
	})

	t.Run("ExponentialBackoffTiming", func(t *testing.T) {
		callCount := 0
		attemptTimes := []time.Time{}

		// Create a mock read function that always fails
		mockRead := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			callCount++
			attemptTimes = append(attemptTimes, time.Now())
			return diag.Errorf("Error reading vector store file: API error: No file found with id 'file-123' in vector store 'vs-456'")
		}

		simulateRetryLogic(mockRead, 5)

		// Verify exponential backoff: 1s, 2s, 4s, 8s
		expectedBackoffs := []time.Duration{
			0 * time.Second,       // First attempt (no wait)
			1 * time.Second,       // Wait 1s before 2nd attempt
			2 * time.Second,       // Wait 2s before 3rd attempt
			4 * time.Second,       // Wait 4s before 4th attempt
			8 * time.Second,       // Wait 8s before 5th attempt
		}

		if len(attemptTimes) != 5 {
			t.Fatalf("Expected 5 attempts, got %d", len(attemptTimes))
		}

		// Check timing between attempts (with tolerance)
		tolerance := 200 * time.Millisecond
		for i := 1; i < len(attemptTimes); i++ {
			actual := attemptTimes[i].Sub(attemptTimes[i-1])
			expected := expectedBackoffs[i]
			diff := actual - expected

			if diff < -tolerance || diff > tolerance {
				t.Errorf("Attempt %d: expected ~%v backoff, got %v (diff: %v)",
					i, expected, actual, diff)
			}
		}
	})
}

// simulateRetryLogic simulates the retry logic from resourceOpenAIVectorStoreFileReadWithRetry
// This is a simplified version for testing purposes
func simulateRetryLogic(readFunc mockReadFunc, maxRetries int) diag.Diagnostics {
	if maxRetries <= 0 {
		return diag.Errorf("maxRetries must be at least 1 for vector store file read retries")
	}

	ctx := context.Background()
	d := &schema.ResourceData{}
	var lastErr diag.Diagnostics

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s, 16s
			backoffDuration := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoffDuration)
		}

		diags := readFunc(ctx, d, nil)
		if diags == nil || !diags.HasError() {
			return diags
		}

		// Check if the error is a "file not found" error, which indicates we should retry
		shouldRetry := false
		for _, diag := range diags {
			if containsRetriableError(diag.Summary) || containsRetriableError(diag.Detail) {
				shouldRetry = true
				break
			}
		}

		if !shouldRetry {
			// If it's not a "file not found" error, return immediately
			return diags
		}

		lastErr = diags
	}

	return lastErr
}

// BenchmarkRetryLogic benchmarks the retry logic performance
func BenchmarkRetryLogic(b *testing.B) {
	mockRead := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		return nil // Immediate success
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simulateRetryLogic(mockRead, 5)
	}
}

// TestRetryLogicEdgeCases tests edge cases
func TestRetryLogicEdgeCases(t *testing.T) {
	t.Run("ZeroMaxRetries", func(t *testing.T) {
		mockRead := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			return nil
		}

		result := simulateRetryLogic(mockRead, 0)

		// Should return error for invalid configuration
		if !result.HasError() {
			t.Error("Expected error for maxRetries=0, got success")
		}
		if !strings.Contains(result[0].Summary, "maxRetries") && !strings.Contains(result[0].Detail, "maxRetries") {
			t.Error("Expected error message to mention maxRetries configuration")
		}
	})

	t.Run("NegativeMaxRetries", func(t *testing.T) {
		mockRead := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			return nil
		}

		result := simulateRetryLogic(mockRead, -1)

		// Should return error for invalid configuration
		if !result.HasError() {
			t.Error("Expected error for maxRetries=-1, got success")
		}
	})

	t.Run("SingleRetry", func(t *testing.T) {
		callCount := 0
		mockRead := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			callCount++
			return diag.Errorf("Error: No file found")
		}

		result := simulateRetryLogic(mockRead, 1)

		// Should make exactly 1 call
		if callCount != 1 {
			t.Errorf("Expected 1 call with maxRetries=1, got %d", callCount)
		}
		if !result.HasError() {
			t.Error("Expected error when all retries fail")
		}
	})

	t.Run("ErrorMessageCaseInsensitivity", func(t *testing.T) {
		testCases := []struct {
			errorMsg    string
			shouldRetry bool
		}{
			{"No file found", true},
			{"no file found", true},          // Now case-insensitive!
			{"FILE NOT FOUND", true},         // Now case-insensitive!
			{"404 Not Found", true},          // HTTP 404 errors
			{"Resource not found", true},
			{"resource not found", true},
			{"Unauthorized", false},          // Non-retriable
			{"Rate limit exceeded", false},   // Non-retriable
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("Error: %s", tc.errorMsg), func(t *testing.T) {
				result := containsRetriableError(tc.errorMsg)
				if result != tc.shouldRetry {
					t.Errorf("Expected shouldRetry=%v for '%s', got %v",
						tc.shouldRetry, tc.errorMsg, result)
				}
			})
		}
	})
}
