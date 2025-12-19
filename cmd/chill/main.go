package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: chill <duration>\n")
		fmt.Fprintf(os.Stderr, "Example: chill 15s\n")
		fmt.Fprintf(os.Stderr, "Formats: 5s, 2m, 1h30m, etc.\n")
		os.Exit(1)
	}

	duration, err := time.ParseDuration(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid duration: %v\n", err)
		os.Exit(1)
	}

	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	// Create a timer
	timer := time.NewTimer(duration)
	defer timer.Stop()

	// Channel to signal skip and cleanup completion
	skipChan := make(chan struct{})
	cleanupDone := make(chan struct{})

	fmt.Printf("Chilling for %v (press Ctrl+G to skip, Ctrl+C to terminate)...\n", duration)

	// Start goroutine to listen for Ctrl+G
	go listenForSkip(skipChan, cleanupDone)

	// Wait for either timer, signal, or skip
	select {
	case <-timer.C:
		fmt.Println("Done chilling!")
	case <-sigChan:
		fmt.Println("\nTerminated.")
		os.Exit(0)
	case <-skipChan:
		fmt.Println("\nChilling skipped.")
		// Wait for goroutine to finish cleanup before exiting
		<-cleanupDone
	}
}

func listenForSkip(skipChan chan<- struct{}, cleanupDone chan<- struct{}) {
	// Save original terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// If we can't set raw mode, just return (might not be a TTY)
		close(cleanupDone)
		return
	}
	defer func() {
		term.Restore(int(os.Stdin.Fd()), oldState)
		close(cleanupDone)
	}()

	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}

		// Ctrl+G is ASCII 7
		if buf[0] == 7 {
			skipChan <- struct{}{}
			return
		}
	}
}
