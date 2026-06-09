package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// testRetryBaseDelay keeps the exponential backoff measurable but fast in tests.
const testRetryBaseDelay = 20 * time.Millisecond

// TestRetryLogicBehavior tests the actual retry behavior by simulating different scenarios
func TestRetryLogicBehavior(t *testing.T) {
	ctx := context.Background()

	t.Run("SuccessOnFirstAttempt", func(t *testing.T) {
		callCount := 0
		startTime := time.Now()

		err := retryVectorStoreFileRead(ctx, 5, testRetryBaseDelay, func() error {
			callCount++
			return nil // Success on first try
		})

		elapsed := time.Since(startTime)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if callCount != 1 {
			t.Errorf("Expected 1 call, got %d", callCount)
		}
		// Should complete almost immediately (no retries)
		if elapsed > 100*time.Millisecond {
			t.Errorf("Expected to complete quickly, took %v", elapsed)
		}
	})

	t.Run("SuccessAfterThreeRetries", func(t *testing.T) {
		callCount := 0
		startTime := time.Now()

		// Fails twice with a retriable error, then succeeds on the 3rd attempt
		err := retryVectorStoreFileRead(ctx, 5, testRetryBaseDelay, func() error {
			callCount++
			if callCount < 3 {
				return fmt.Errorf("Error reading vector store file: API error: No file found with id 'file-123' in vector store 'vs-456'")
			}
			return nil
		})

		elapsed := time.Since(startTime)

		if err != nil {
			t.Errorf("Expected no error after retries, got: %v", err)
		}
		if callCount != 3 {
			t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount)
		}
		// Should have waited: baseDelay + 2*baseDelay = 3*baseDelay (approximately)
		expected := 3 * testRetryBaseDelay
		if elapsed < expected || elapsed > expected+200*time.Millisecond {
			t.Errorf("Expected ~%v elapsed (1x + 2x backoff), got %v", expected, elapsed)
		}
	})

	t.Run("MaxRetriesExhausted", func(t *testing.T) {
		callCount := 0
		maxRetries := 5

		err := retryVectorStoreFileRead(ctx, maxRetries, testRetryBaseDelay, func() error {
			callCount++
			return fmt.Errorf("Error reading vector store file: API error: No file found with id 'file-123' in vector store 'vs-456'")
		})

		if err == nil {
			t.Error("Expected error after max retries exhausted")
		}
		if callCount != maxRetries {
			t.Errorf("Expected %d calls (max retries), got %d", maxRetries, callCount)
		}
	})

	t.Run("NonRetriableErrorFailsImmediately", func(t *testing.T) {
		callCount := 0
		startTime := time.Now()

		err := retryVectorStoreFileRead(ctx, 5, testRetryBaseDelay, func() error {
			callCount++
			return fmt.Errorf("Error reading vector store file: API error: Unauthorized")
		})

		elapsed := time.Since(startTime)

		if err == nil {
			t.Error("Expected error for unauthorized")
		}
		if callCount != 1 {
			t.Errorf("Expected 1 call (no retries for non-retriable error), got %d", callCount)
		}
		// Should fail immediately without waiting
		if elapsed > 100*time.Millisecond {
			t.Errorf("Expected to fail immediately, took %v", elapsed)
		}
	})

	t.Run("ExponentialBackoffTiming", func(t *testing.T) {
		attemptTimes := []time.Time{}

		_ = retryVectorStoreFileRead(ctx, 5, testRetryBaseDelay, func() error {
			attemptTimes = append(attemptTimes, time.Now())
			return fmt.Errorf("Error reading vector store file: API error: No file found with id 'file-123' in vector store 'vs-456'")
		})

		// Verify exponential backoff: 1x, 2x, 4x, 8x baseDelay
		expectedBackoffs := []time.Duration{
			0,                      // First attempt (no wait)
			1 * testRetryBaseDelay, // Wait 1x before 2nd attempt
			2 * testRetryBaseDelay, // Wait 2x before 3rd attempt
			4 * testRetryBaseDelay, // Wait 4x before 4th attempt
			8 * testRetryBaseDelay, // Wait 8x before 5th attempt
		}

		if len(attemptTimes) != 5 {
			t.Fatalf("Expected 5 attempts, got %d", len(attemptTimes))
		}

		// Check timing between attempts: at least the expected backoff,
		// with generous upper tolerance to avoid flakiness on slow machines
		tolerance := 150 * time.Millisecond
		for i := 1; i < len(attemptTimes); i++ {
			actual := attemptTimes[i].Sub(attemptTimes[i-1])
			expected := expectedBackoffs[i]

			if actual < expected || actual > expected+tolerance {
				t.Errorf("Attempt %d: expected ~%v backoff, got %v", i, expected, actual)
			}
		}
	})
}

// BenchmarkRetryLogic benchmarks the retry logic performance
func BenchmarkRetryLogic(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = retryVectorStoreFileRead(ctx, 5, testRetryBaseDelay, func() error {
			return nil // Immediate success
		})
	}
}

// TestRetryLogicEdgeCases tests edge cases
func TestRetryLogicEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("ZeroMaxRetries", func(t *testing.T) {
		err := retryVectorStoreFileRead(ctx, 0, testRetryBaseDelay, func() error {
			return nil
		})

		// Should return error for invalid configuration
		if err == nil {
			t.Error("Expected error for maxRetries=0, got success")
		}
		if err != nil && !strings.Contains(err.Error(), "maxRetries") {
			t.Error("Expected error message to mention maxRetries configuration")
		}
	})

	t.Run("NegativeMaxRetries", func(t *testing.T) {
		err := retryVectorStoreFileRead(ctx, -1, testRetryBaseDelay, func() error {
			return nil
		})

		// Should return error for invalid configuration
		if err == nil {
			t.Error("Expected error for maxRetries=-1, got success")
		}
	})

	t.Run("SingleRetry", func(t *testing.T) {
		callCount := 0
		err := retryVectorStoreFileRead(ctx, 1, testRetryBaseDelay, func() error {
			callCount++
			return fmt.Errorf("Error: No file found")
		})

		// Should make exactly 1 call
		if callCount != 1 {
			t.Errorf("Expected 1 call with maxRetries=1, got %d", callCount)
		}
		if err == nil {
			t.Error("Expected error when all retries fail")
		}
	})

	t.Run("ErrorMessageCaseInsensitivity", func(t *testing.T) {
		testCases := []struct {
			errorMsg    string
			shouldRetry bool
		}{
			{"No file found", true},
			{"no file found", true},  // Case-insensitive
			{"FILE NOT FOUND", true}, // Case-insensitive
			{"404 Not Found", true},  // HTTP 404 errors
			{"Resource not found", true},
			{"resource not found", true},
			{"Unauthorized", false},        // Non-retriable
			{"Rate limit exceeded", false}, // Non-retriable
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
