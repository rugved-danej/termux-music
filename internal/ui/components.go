package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func makeDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(colorGreen).Bold(true).
		BorderLeftForeground(colorGreen)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(colorYellow).
		BorderLeftForeground(colorGreen)
	d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(colorWhite)
	d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(colorGray)
	return d
}

func makeList(title string) list.Model {
	l := list.New([]list.Item{}, makeDelegate(), 0, 0)
	l.Title = title
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(colorPurple).Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(colorDim).
		MarginBottom(1)
	return l
}
func (m model) renderHeader() string {
	logoText := "TERMUX MUSIC"
	var logo string
	colors := []lipgloss.Color{colorGreen, colorCyan, colorPurple, colorPink, colorYellow}
	for i, c := range logoText {
		colorIdx := ((m.animFrame / 2) + i) % len(colors)
		if string(c) == " " {
			logo += " "
		} else {
			logo += lipgloss.NewStyle().Foreground(colors[colorIdx]).Bold(true).Render(string(c))
		}
	}
	status := ""
	if m.statusMsg != "" {
		status = "  " + lipgloss.NewStyle().Foreground(colorYellow).Italic(true).Render(m.statusMsg)
	}
	content := logo + status

	if bp(m.width) == "narrow" {
		return lipgloss.NewStyle().Padding(0, 1).Render(content)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPurple).
		Padding(0, 2).
		Width(clamp(m.width-4, 10, 9999)).
		Render(content)
}

func (m model) renderTabBar() string {
	type tab struct {
		short string
		long  string
		state int
	}
	tabs := []tab{
		{"[1]S", "[1] Search", stateSearch},
		{"[2]T", "[2] Trending", stateTrending},
		{"[3]L", "[3] Library", stateLibrary},
		{"[4]P", "[4] Playlists", statePlaylists},
		{"[5]H", "[5] History", stateHistory},
		{"[6]Q", "[6] Queue", stateQueue},
		{"[7]D", "[7] Discover", stateDiscover},
	}

	narrow := bp(m.width) == "narrow"
	var parts []string
	for _, t := range tabs {
		active := m.state == t.state ||
			(t.state == stateSearch && (m.state == stateLoading || m.state == stateResults)) ||
			(t.state == stateTrending && m.state == stateTrendingLoad) ||
			(t.state == stateDiscover && m.state == stateDiscoverLoad) ||
			(t.state == statePlaylists && (m.state == statePlaylistView ||
				m.state == stateCreatePlaylist || m.state == stateAddToPlaylist))
		label := t.long
		if narrow {
			label = t.short
		}
		if active {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(colorGreen).Bold(true).
				Padding(0, 1).Render(label))
		} else {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(colorGray).Padding(0, 1).Render(label))
		}
	}

	sep := "  "
	if narrow {
		sep = " "
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(colorDim).
		MarginTop(1).MarginBottom(1).
		Render(strings.Join(parts, sep))
}

func (m model) renderMiniPlayer() string {
	if m.playing.ID == "" || m.state == statePlaying || m.width < 50 {
		return ""
	}
	icon := lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render("▶")
	if m.paused {
		icon = lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render("⏸")
	}
	maxT := clamp(m.width-36, 10, 9999)
	title := truncate(m.playing.Title, maxT)
	hint := lipgloss.NewStyle().Foreground(colorGray).Render("  [Space]  [n] Skip  [Esc] Open")

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(colorDim).
		Padding(0, 1).
		Width(clamp(m.width-4, 10, 9999)).
		Render(icon + "  " +
			lipgloss.NewStyle().Foreground(colorWhite).Render(title) + hint)
}

func (m model) renderControls(lines []string) string {
	if bp(m.width) == "narrow" {
		return lipgloss.NewStyle().Foreground(colorGray).MarginTop(1).
			Render(strings.Join(lines, "\n"))
	}
	all := strings.Join(lines, "  ")
	return lipgloss.NewStyle().Foreground(colorGray).MarginTop(1).Render(all)
}

func customProgressBar(width int, pct float64, animFrame int, paused bool) string {
	filledW := int(float64(width) * pct)
	if filledW < 0 {
		filledW = 0
	}
	if filledW > width {
		filledW = width
	}
	emptyW := width - filledW

	colors := []lipgloss.Color{colorCyan, colorPurple, colorPink}
	bar := ""

	for i := 0; i < filledW; i++ {
		cIdx := ((i + animFrame/2) % 30) / 10
		if cIdx >= len(colors) {
			cIdx = 0
		}
		if paused {
			bar += lipgloss.NewStyle().Foreground(colorDim).Render("█")
		} else {
			bar += lipgloss.NewStyle().Foreground(colors[cIdx]).Render("█")
		}
	}

	if emptyW > 0 {
		bar += lipgloss.NewStyle().Foreground(colorDim).Render(strings.Repeat("░", emptyW))
	}
	return bar + " "
}
