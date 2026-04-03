package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// eventMap maps BUSD event names to notification titles and the payload keys
// used as subtitle (first match wins).
var eventMap = map[string]struct {
	title      string
	payloadKey []string
}{
	"pipeline.run.completed":    {"Pipeline done", []string{"pipeline", "name"}},
	"pipeline.run.failed":       {"Pipeline failed", []string{"pipeline", "name"}},
	"pipeline.step.failed":      {"Step failed", []string{"step", "name"}},
	"agent.run.completed":       {"Agent done", []string{"agent", "name"}},
	"agent.run.failed":          {"Agent failed", []string{"agent", "name"}},
	"agent.run.clarification":   {"gl1tch needs input", []string{"agent", "name"}},
	"cron.job.completed":        {"Cron done", []string{"job", "name"}},
	"cron.job.started":          {"Cron started", []string{"job", "name"}},
	"game.achievement.unlocked": {"Achievement!", []string{"achievement", "name"}},
	"game.bounty.completed":     {"Bounty complete", []string{"bounty", "name"}},
}

// sendNotification fires a macOS notification via osascript.
// subtitle may be empty; when empty the subtitle argument is omitted entirely.
func sendNotification(title, subtitle string) {
	// Escape single quotes for safe inline AppleScript string interpolation.
	safeTitle := escapeSingleQuotes(title)
	safeSubtitle := escapeSingleQuotes(subtitle)

	var script string
	if safeSubtitle != "" {
		script = fmt.Sprintf(
			`display notification "" with title 'gl1tch' subtitle '%s — %s'`,
			safeSubtitle, safeTitle,
		)
	} else {
		script = fmt.Sprintf(
			`display notification "" with title 'gl1tch' subtitle '%s'`,
			safeTitle,
		)
	}

	cmd := exec.Command("osascript", "-e", script)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "gl1tch-notify: osascript: %v\n", err)
	}
}

// escapeSingleQuotes escapes single-quote characters for AppleScript string literals.
func escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}

// handleEvent maps a BUSD event frame to a notification.
// Returns false if the event is not in the allow-list and should be ignored.
func handleEvent(eventName string, payload map[string]any) bool {
	info, ok := eventMap[eventName]
	if !ok {
		return false
	}

	subtitle := ""
	for _, key := range info.payloadKey {
		if v, ok := payload[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				subtitle = s
				break
			}
		}
	}

	sendNotification(info.title, subtitle)
	return true
}
