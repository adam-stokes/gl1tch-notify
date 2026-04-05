package main

/*
#cgo LDFLAGS: -framework Foundation -framework UserNotifications
#include "notify_darwin.h"
*/
import "C"

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
	"mattermost.mention":        {"Mattermost mention", []string{"sender_name", "message"}},
	"mattermost.direct":         {"Mattermost DM", []string{"sender_name", "message"}},
}

// sendNotification delivers a macOS notification via UNUserNotificationCenter.
// Being called from within the app process avoids the osascript daemon-context
// limitation where display notification is silently dropped.
func sendNotification(title, subtitle string) {
	C.sendUserNotification(C.CString(title), C.CString(subtitle))
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
