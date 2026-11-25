package reasoner

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fadikaba81/ObservabilityAgentAI/pkg/models"
)

const (
	maxLogLineLen = 400
	maxPromptLen  = 60000
)

// Bundle is everything collected for one alert, ready to be turned into a prompt.
type Bundle struct {
	Alert    models.Alert
	Start    time.Time
	End      time.Time
	Metrics  []map[string]any
	Traces   []map[string]any
	Logs     []models.Event
	Warnings []string
}

var secretPattern = regexp.MustCompile(`(?i)(password|passwd|secret|token|api[_-]?key|authorization)(["']?\s*[:=]\s*["']?)[^\s"',}]+`)

// Redact masks values that look like credentials before text leaves the process.
func Redact(s string) string {
	return secretPattern.ReplaceAllString(s, "${1}${2}[REDACTED]")
}

// BuildPrompt renders the bundle as plain text for the model.
func BuildPrompt(b Bundle) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "ALERT\nname: %s\nseverity: %s\nservice: %s\nmessage: %s\ndescription: %s\nerror: %s\n",
		b.Alert.Name, b.Alert.Severity, b.Alert.Service, b.Alert.Message, b.Alert.Description, b.Alert.ErrorMessage)
	fmt.Fprintf(&sb, "window: %s to %s (UTC)\n\n",
		b.Start.UTC().Format(time.RFC3339), b.End.UTC().Format(time.RFC3339))

	sb.WriteString("NEW RELIC METRICS (window summary)\n")
	writeRows(&sb, b.Metrics)

	sb.WriteString("\nNEW RELIC SLOWEST TRACES\n")
	writeRows(&sb, b.Traces)

	fmt.Fprintf(&sb, "\nSUMO LOGS (%d lines, oldest first)\n", len(b.Logs))
	for _, l := range b.Logs {
		if sb.Len() > maxPromptLen {
			sb.WriteString("[remaining log lines omitted for length]\n")
			break
		}
		fmt.Fprintf(&sb, "%s [%s] %s\n",
			timeOrUnknown(l.Timestamps), l.Severity, truncate(l.RawLog, maxLogLineLen))
	}

	if len(b.Warnings) > 0 {
		sb.WriteString("\nDATA COLLECTION WARNINGS\n")
		for _, w := range b.Warnings {
			fmt.Fprintf(&sb, "- %s\n", w)
		}
	}

	return Redact(sb.String())
}

func writeRows(sb *strings.Builder, rows []map[string]any) {
	if len(rows) == 0 {
		sb.WriteString("(no data)\n")
		return
	}
	for _, row := range rows {
		b, err := json.Marshal(row)
		if err != nil {
			continue
		}
		sb.Write(b)
		sb.WriteString("\n")
	}
}

func timeOrUnknown(t time.Time) string {
	if t.IsZero() {
		return "unknown-time"
	}
	return t.UTC().Format(time.RFC3339)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
