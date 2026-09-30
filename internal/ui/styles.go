package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorGreen  = lipgloss.Color("#00FF7F")
	colorPurple = lipgloss.Color("#CC55FF")
	colorYellow = lipgloss.Color("#FFD700")
	colorCyan   = lipgloss.Color("#00FFFF")
	colorOrange = lipgloss.Color("#FF8C00")
	colorPink   = lipgloss.Color("#FF69B4")
	colorGray   = lipgloss.Color("#555555")
	colorDim    = lipgloss.Color("#2A2A2A")
	colorWhite  = lipgloss.Color("#EEEEEE")
)

func bp(w int) string {
	if w < 60 {
		return "narrow"
	}
	if w < 100 {
		return "medium"
	}
	return "wide"
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func progressWidth(w int) int {
	switch bp(w) {
	case "narrow":
		return clamp(w-8, 5, 9999)
	case "medium":
		return clamp(w-12, 5, 9999)
	default:
		return clamp(w-14, 5, 9999)
	}
}

func listHeight(h int) int {
	return clamp(h-12, 4, 9999)
}

func divider(w int) string {
	if w <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Foreground(colorDim).Render(strings.Repeat("─", w))
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
