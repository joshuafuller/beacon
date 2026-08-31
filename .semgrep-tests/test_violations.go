// Test file to verify Semgrep rules catch violations
// This file intentionally contains rule violations to test Semgrep detection
//
// Run: semgrep --config=.semgrep.yml .semgrep-tests/
package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// ============================================================================
// F-4: Timer Leaks (beacon-timer-leak)
// ============================================================================

func BadTimerNoStop() {
	// SHOULD TRIGGER: beacon-timer-leak
	timer := time.NewTimer(5 * time.Second)
	// Missing: defer timer.Stop()
	<-timer.C
}

func GoodTimerWithStop() {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop() // Correct!
	<-timer.C
}

// ============================================================================
// F-4: Ticker Leaks (beacon-ticker-leak)
// ============================================================================

func BadTickerNoStop() {
	// SHOULD TRIGGER: beacon-ticker-leak
	ticker := time.NewTicker(1 * time.Second)
	// Missing: defer ticker.Stop()
	for range ticker.C {
		break
	}
}

func GoodTickerWithStop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop() // Correct!
	for range ticker.C {
		break
	}
}

// ============================================================================
// F-4: WaitGroup Missing Done (beacon-waitgroup-missing-done)
// ============================================================================

func BadWaitGroupMissingDone() {
	var wg sync.WaitGroup

	// SHOULD TRIGGER: beacon-waitgroup-missing-done
	wg.Add(1)
	go func() {
		// Missing: defer wg.Done()
		time.Sleep(1 * time.Second)
	}()

	wg.Wait()
}

func GoodWaitGroupWithDone() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done() // Correct!
		time.Sleep(1 * time.Second)
	}()

	wg.Wait()
}

// ============================================================================
// F-4: Unbuffered Result Channel (beacon-unbuffered-result-channel)
// ============================================================================

func BadUnbufferedChannel() {
	// SHOULD TRIGGER: beacon-unbuffered-result-channel
	ch := make(chan string)
	go func() {
		ch <- "result" // Blocks forever if context canceled before receive
	}()
	result := <-ch
	_ = result
}

func GoodBufferedChannel() {
	ch := make(chan string, 1) // Correct!
	go func() {
		ch <- "result"
	}()
	result := <-ch
	_ = result
}

// ============================================================================
// F-4: Mutex Missing Defer Unlock (beacon-mutex-defer-unlock)
// ============================================================================

func BadMutexNoDefer() {
	var mu sync.Mutex

	// SHOULD TRIGGER: beacon-mutex-defer-unlock
	mu.Lock()
	// Missing: defer mu.Unlock()
	// Do work
	mu.Unlock() // Manual unlock - risky if panic occurs
}

func GoodMutexWithDefer() {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock() // Correct!
	// Do work
}

// ============================================================================
// F-3: Error Message Punctuation (beacon-error-message-punctuation)
// ============================================================================

func BadErrorPunctuation() error {
	// SHOULD TRIGGER: beacon-error-message-punctuation
	return errors.New("Something went wrong.")
}

func GoodErrorNoPunctuation() error {
	return errors.New("something went wrong") // Correct!
}

func BadErrorfPunctuation() error {
	// SHOULD TRIGGER: beacon-error-message-punctuation
	return fmt.Errorf("Failed to connect!")
}

func GoodErrorfNoPunctuation() error {
	return fmt.Errorf("failed to connect") // Correct!
}

// ============================================================================
// F-3: Error Log and Return (beacon-error-log-and-return)
// ============================================================================

func BadLogAndReturn(data []byte) error {
	// SHOULD TRIGGER: beacon-error-log-and-return
	err := processData(data)
	if err != nil {
		fmt.Printf("Error processing: %v\n", err)
		return err // Duplicate logging when caller also logs
	}
	return nil
}

func GoodReturnOnly(data []byte) error {
	// Correct! Let caller decide to log
	return processData(data)
}

func processData(data []byte) error {
	return nil
}

// ============================================================================
// F-7: File Missing Defer Close (beacon-file-missing-defer-close)
// ============================================================================

func BadFileMissingDefer() error {
	// SHOULD TRIGGER: beacon-file-missing-defer-close
	file, err := os.Open("test.txt")
	if err != nil {
		return err
	}
	// Missing: defer file.Close()
	_ = file
	return nil
}

func GoodFileWithDefer() error {
	file, err := os.Open("test.txt")
	if err != nil {
		return err
	}
	defer file.Close() // Correct!
	_ = file
	return nil
}

// ============================================================================
// RFC 6762: Hardcoded mDNS Constants (beacon-hardcoded-mdns-port)
// ============================================================================

func BadHardcodedPort() string {
	// SHOULD TRIGGER: beacon-hardcoded-mdns-port
	return ":5353"
}

func BadHardcodedMulticastAddr() string {
	// SHOULD TRIGGER: beacon-hardcoded-multicast-address
	return "224.0.0.251"
}

// Good: Use constants instead
// const DefaultPort = 5353
// const DefaultMulticastIPv4 = "224.0.0.251"

// ============================================================================
// F-4: Context Not First Param (beacon-context-not-first-param)
// ============================================================================

// SHOULD TRIGGER: beacon-context-not-first-param
func BadContextNotFirst(name string, ctx context.Context) error {
	_ = name
	<-ctx.Done()
	return ctx.Err()
}

func GoodContextFirst(ctx context.Context, name string) error {
	_ = name
	<-ctx.Done()
	return ctx.Err()
}

// ============================================================================
// F-4: Context Not Checked in Loop (beacon-context-not-checked-loop)
// ============================================================================

func BadLoopNoContextCheck(ctx context.Context) {
	// SHOULD TRIGGER: beacon-context-not-checked-loop
	for {
		time.Sleep(1 * time.Second)
		// No check for ctx.Done()
	}
}

func GoodLoopWithContextCheck(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return // Correct!
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// ============================================================================
// Constitution: Error Swallowing (beacon-error-swallowing)
// ============================================================================

func BadErrorSwallowing() {
	// SHOULD TRIGGER: beacon-error-swallowing
	_, _ = os.Open("test.txt") // Error ignored with _
}

func GoodErrorHandling() error {
	_, err := os.Open("test.txt")
	if err != nil {
		return err // Correct!
	}
	return nil
}

// ============================================================================
// Constitution: Error Capitalization (beacon-error-capitalization)
// ============================================================================

func BadErrorCapitalized() error {
	// SHOULD TRIGGER: beacon-error-capitalization
	return fmt.Errorf("Failed to process request", "arg")
}

func GoodErrorLowercase() error {
	return fmt.Errorf("failed to process request: %s", "arg") // Correct!
}
