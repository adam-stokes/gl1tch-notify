package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const (
	registrationMsg = `{"name":"gl1tch-notify","subscribe":["*"]}` + "\n"
	backoffMin      = 1 * time.Second
	backoffMax      = 30 * time.Second
)

// socketPath returns the BUSD unix socket path, trying XDG_RUNTIME_DIR first,
// then $HOME/.cache/glitch/bus.sock.
func socketPath() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, "glitch", "bus.sock")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".cache", "glitch", "bus.sock")
}

// eventFrame is the wire format for events received from the bus.
type eventFrame struct {
	Event   string         `json:"event"`
	Payload map[string]any `json:"payload"`
}

// busLoop connects to the BUSD socket, reads events, and dispatches them.
// It sends connection-state booleans to statusCh and retries with exponential
// backoff on disconnect.
func busLoop(statusCh chan<- bool) {
	backoff := backoffMin
	path := socketPath()

	for {
		conn, err := net.Dial("unix", path)
		if err != nil {
			// Socket not present means gl1tch isn't running — silent retry.
			if !errors.Is(err, syscall.ENOENT) {
				fmt.Fprintf(os.Stderr, "gl1tch-notify: connect %s: %v (retry in %s)\n", path, err, backoff)
			}
			statusCh <- false
			time.Sleep(backoff)
			backoff = min(backoff*2, backoffMax)
			continue
		}

		// Register
		if _, err := fmt.Fprint(conn, registrationMsg); err != nil {
			fmt.Fprintf(os.Stderr, "gl1tch-notify: register: %v\n", err)
			conn.Close()
			statusCh <- false
			time.Sleep(backoff)
			continue
		}

		backoff = backoffMin // reset on successful connect
		statusCh <- true

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			var frame eventFrame
			if err := json.Unmarshal(line, &frame); err != nil {
				fmt.Fprintf(os.Stderr, "gl1tch-notify: parse: %v\n", err)
				continue
			}
			handleEvent(frame.Event, frame.Payload)
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "gl1tch-notify: read: %v\n", err)
		}
		conn.Close()
		statusCh <- false

		fmt.Fprintf(os.Stderr, "gl1tch-notify: disconnected, retry in %s\n", backoff)
		time.Sleep(backoff)
		backoff = min(backoff*2, backoffMax)
	}
}

// min returns the smaller of two durations.
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
