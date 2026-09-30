package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	margin := lipgloss.NewStyle().Margin(1, 1)

	switch m.state {
	case statePlaying:
		return m.nowPlayingView()
	}

	header := m.renderHeader()
	tabs := m.renderTabBar()
	mini := m.renderMiniPlayer()

	switch m.state {

	case stateSearch:
		searchBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2).
			MarginTop(1).
			Width(clamp(m.width-4, 10, 9999)).
			Render(m.textInput.View())
		hint := lipgloss.NewStyle().Foreground(colorGray).MarginTop(1).MarginLeft(2).
			Render("Enter to search  •  Ctrl+C to quit")
		return margin.Render(header + "\n" + tabs + searchBox + "\n" + hint + "\n" + mini)

	case stateLoading, stateTrendingLoad, stateDiscoverLoad:
		label := fmt.Sprintf("Searching for \"%s\"...", m.query)
		if m.state == stateTrendingLoad {
			label = "Fetching trending music..."
		} else if m.state == stateDiscoverLoad {
			label = "Building your personalized recommendations..."
		}
		spinner := lipgloss.NewStyle().Margin(2, 1).
			Render(m.spinner.View() + "  " +
				lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render(label))
		return margin.Render(header + "\n" + tabs + spinner)

	case stateResults, stateTrending, stateLibrary, stateHistory, statePlaylistView, stateDiscover:
		helpMap := map[int][]string{
			stateResults:      {"[Enter] Play", "[q] Queue", "[f] Fav", "[d] DL", "[a] +Playlist", "[/] Filter", "[1-7] Tab"},
			stateTrending:     {"[Enter] Play", "[q] Queue", "[f] Fav", "[d] DL", "[a] +Playlist", "[/] Filter", "[1-7] Tab"},
			stateDiscover:     {"[Enter] Play", "[q] Queue", "[f] Fav", "[d] DL", "[a] +Playlist", "[/] Filter", "[1-7] Tab"},
			stateLibrary:      {"[Enter] Play", "[q] Queue", "[f] Fav", "[d] DL", "[a] +Playlist", "[i] Import", "[/] Filter", "[1-7] Tab"},
			stateHistory:      {"[Enter] Play", "[q] Queue", "[f] Fav", "[a] +Playlist", "[/] Filter", "[1-7] Tab"},
			statePlaylistView: {"[Enter] Play", "[q] Queue", "[f] Fav", "[d] DL", "[x] Remove", "[/] Filter", "[Esc] Back"},
		}
		return margin.Render(header + "\n" + tabs + m.list.View() + "\n" +
			m.renderControls(helpMap[m.state]) + "\n" + mini)

	case stateQueue:
		return margin.Render(header + "\n" + tabs + m.list.View() + "\n" +
			m.renderControls([]string{"[Enter] Play", "[x] Remove", "[J/K] Move", "[1-7] Tab"}) + "\n" + mini)

	case stateDownload:
		return margin.Render(header + "\n" + tabs + m.list.View() + "\n" +
			m.renderControls([]string{"[Enter] Download", "[Esc] Cancel"}) + "\n" + mini)

	case statePlaylists:
		return margin.Render(header + "\n" + tabs + m.list.View() + "\n" +
			m.renderControls([]string{"[Enter] Open", "[c] Create", "[x] Delete", "[1-7] Tab"}) +
			"\n" + mini)

	case stateCreatePlaylist:
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2).
			MarginTop(1).
			Width(clamp(m.width-4, 10, 9999)).
			Render(m.createPlaylistInput.View())

		title := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).MarginLeft(2).Render("Create New Playlist")
		hint := lipgloss.NewStyle().Foreground(colorGray).MarginTop(1).MarginLeft(2).Render("Enter to save  •  Esc to cancel")

		return margin.Render(header + "\n" + tabs + "\n" + title + "\n" + inputBox + "\n" + hint + "\n" + mini)

	case stateAddToPlaylist:
		return margin.Render(header + "\n" + tabs + m.list.View() + "\n" +
			m.renderControls([]string{"[Enter] Add to playlist", "[Esc] Cancel"}))
	}
	return ""
}
