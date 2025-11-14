package utility

import (
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	Indent              = "  "
	DoubleIndent        = Indent + Indent
	DefaultSnippetLimit = 160
	dividerWidth        = 72
)

var dividerLine = strings.Repeat("-", dividerWidth)

func DividerLine() string {
	return dividerLine
}

func StatusBadge(success bool) string {
	if success {
		return color.GreenString("[PASS]")
	}
	return color.RedString("[FAIL]")
}

func DurationBadge(duration time.Duration) string {
	switch {
	case duration < 500*time.Millisecond:
		return color.GreenString("%s", duration)
	case duration < 2*time.Second:
		return color.YellowString("%s", duration)
	default:
		return color.RedString("%s", duration)
	}
}

func FormatSnippet(body string, emptyPlaceholder string, maxLen int) string {
	snippet := strings.TrimSpace(body)
	if snippet == "" {
		if emptyPlaceholder != "" {
			return emptyPlaceholder
		}
		return ""
	}
	limit := maxLen
	if limit <= 0 {
		limit = DefaultSnippetLimit
	}
	if len(snippet) > limit {
		if limit > 3 {
			return snippet[:limit-3] + "..."
		}
		return snippet[:limit]
	}
	return snippet
}
