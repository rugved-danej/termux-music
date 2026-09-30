package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/rugved-danej/termux-music/internal/core"
	"github.com/rugved-danej/termux-music/internal/models"
)

type (
	searchResultMsg []models.Song
	trendingMsg     []models.Song
	artistRadioMsg  []models.Song
	lyricsReadyMsg  []models.LyricLine
	songFinishedMsg struct{}
	downloadDoneMsg struct{ path string }
	errMsg          string
	tickMsg         time.Time
	animTickMsg     time.Time
	discoverMsg     []models.Song
)

const (
	stateSearch         = 0
	stateLoading        = 1
	stateResults        = 2
	statePlaying        = 3
	stateLibrary        = 4
	stateTrending       = 5
	stateTrendingLoad   = 6
	statePlaylists      = 7
	statePlaylistView   = 8
	stateHistory        = 9
	stateCreatePlaylist = 10
	stateAddToPlaylist  = 11
	stateQueue          = 12
	stateDownload       = 13
	stateDiscoverLoad   = 14
	stateDiscover       = 15

	repeatOff = 0
	repeatAll = 1
	repeatOne = 2
)

var sleepOptions = []int{0, 15, 30, 60, 120}

type model struct {
	textInput           textinput.Model
	createPlaylistInput textinput.Model
	list                list.Model
	spinner             spinner.Model
	progress            progress.Model

	state  int
	width  int
	height int

	query   string
	library models.Library

	playing    models.Song
	queue      []models.Song
	radioOn    bool
	shuffle    bool
	repeatMode int
	speed      float64

	mpvPos    float64
	paused    bool
	volume    int
	scrobbled bool

	lyrics      []models.LyricLine
	lyricOffset int
	eqIndex     int

	sleepIdx int
	sleepEnd time.Time

	addTargetSong    models.Song
	selectedPlaylist string
	downloading      bool
	statusMsg        string

	downloadTarget models.Song
	prevState      int
	prevList       list.Model
	trackArmed     bool
	animFrame      int
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type a song, artist or album..."
	ti.Prompt = " ❯ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(colorWhite)
	ti.Focus()
	ti.CharLimit = 150

	cpi := textinput.New()
	cpi.Placeholder = "Enter playlist name..."
	cpi.Prompt = " ❯ "
	cpi.PromptStyle = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	cpi.TextStyle = lipgloss.NewStyle().Foreground(colorWhite)
	cpi.CharLimit = 50

	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(colorPurple)

	p := progress.New(
		progress.WithGradient("#00FF7F", "#CC55FF"),
		progress.WithoutPercentage(),
	)
	p.Width = 40

	return model{
		textInput:           ti,
		createPlaylistInput: cpi,
		list:                makeList("Search Results"),
		spinner:             s,
		progress:            p,
		library:             core.LoadLibrary(),
		state:               stateSearch,
		volume:              100,
		eqIndex:             0,
		speed:               1.0,
	}
}
