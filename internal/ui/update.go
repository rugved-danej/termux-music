package ui

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/rugved-danej/termux-music/internal/core"
	"github.com/rugved-danej/termux-music/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, animTickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = progressWidth(m.width)
		lh := listHeight(m.height)
		lw := clamp(m.width-4, 10, 9999)
		m.list.SetSize(lw, lh)
		m.list, _ = m.list.Update(msg)

	case tea.KeyMsg:
		key := msg.String()

		isListState := m.state == stateResults || m.state == stateTrending ||
			m.state == stateLibrary || m.state == statePlaylists ||
			m.state == statePlaylistView || m.state == stateHistory ||
			m.state == stateAddToPlaylist || m.state == stateQueue || m.state == stateDownload ||
			m.state == stateDiscover

		if isListState && m.list.FilterState() == list.Filtering {
			if key == "ctrl+c" {
				if m.state == statePlaying || m.mpvPos > 0 {
					m.playing.LastPos = m.mpvPos
					m.library.Upsert(m.playing)
					core.SaveLibrary(m.library)
				}
				core.ClearNotification()
				core.ReleaseWakeLock()
				exec.Command("killall", "mpv").Run()
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == stateCreatePlaylist {
			switch key {
			case "enter":
				name := strings.TrimSpace(m.createPlaylistInput.Value())
				if name != "" {
					m.library.CreatePlaylist(name)
					core.SaveLibrary(m.library)
					m.statusMsg = "Created: " + name
				}
				m.createPlaylistInput.SetValue("")
				m.state = statePlaylists
				m.loadPlaylistsList()
				return m, nil
			case "esc":
				m.createPlaylistInput.SetValue("")
				m.state = statePlaylists
				return m, nil
			default:
				var cmd tea.Cmd
				m.createPlaylistInput, cmd = m.createPlaylistInput.Update(msg)
				return m, cmd
			}
		}

		if m.state == stateAddToPlaylist {
			switch key {
			case "enter":
				if i, ok := m.list.SelectedItem().(playlistItem); ok {
					m.library.Upsert(m.addTargetSong)
					m.library.AddSongToPlaylist(i.pl.Name, m.addTargetSong.ID)
					core.SaveLibrary(m.library)
					m.statusMsg = "Added to: " + i.pl.Name
					m.state = stateResults
				}
				return m, nil
			case "esc":
				m.state = stateResults
				return m, nil
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch key {
		case "ctrl+c":
			if m.state == statePlaying || m.mpvPos > 0 {
				m.playing.LastPos = m.mpvPos
				m.library.Upsert(m.playing)
				core.SaveLibrary(m.library)
			}
			core.ClearNotification()
			core.ReleaseWakeLock()
			exec.Command("killall", "mpv").Run()
			return m, tea.Quit

		case "esc":
			if m.state == stateDownload {
				m.state = m.prevState
				m.list = m.prevList
				return m, nil
			}
			if m.state == statePlaying {
				if len(m.query) > 0 {
					m.state = stateResults
				} else {
					m.state = stateSearch
					m.textInput.Focus()
				}
				return m, nil
			}
			if m.state != stateSearch {
				m.state = stateSearch
				m.textInput.Focus()
				return m, nil
			}
			if m.state == stateSearch && m.playing.ID != "" {
				m.state = statePlaying
				m.textInput.Blur()
				return m, nil
			}

		case "1":
			m.state = stateSearch
			m.textInput.Focus()
			return m, nil
		case "2":
			m.state = stateTrendingLoad
			m.list = makeList("Trending Music")
			m.list.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
			return m, tea.Batch(m.spinner.Tick, FetchTrendingCmd())
		case "3":
			m.state = stateLibrary
			m.loadLibraryList()
			return m, nil
		case "4":
			m.state = statePlaylists
			m.loadPlaylistsList()
			return m, nil
		case "5":
			m.state = stateHistory
			m.loadHistoryList()
			return m, nil
		case "6":
			m.state = stateQueue
			m.loadQueueList()
			return m, nil
		case "7":
			m.state = stateDiscoverLoad
			m.list = makeList("Discovering...")
			m.list.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
			return m, tea.Batch(m.spinner.Tick, fetchDiscoverCmd(m.library))

		case "enter":
			switch m.state {
			case stateSearch:
				q := strings.TrimSpace(m.textInput.Value())
				if q != "" {
					m.query = q
					m.state = stateLoading
					m.textInput.Blur()
					return m, tea.Batch(m.spinner.Tick, SearchYoutubeCmd(q))
				}
			case stateResults, stateTrending, stateLibrary, stateHistory, statePlaylistView, stateDiscover:
				if li, ok := m.list.SelectedItem().(listItem); ok {
					items := m.list.Items()
					idx := m.list.Index()
					var newQueue []models.Song

					for i := idx + 1; i < len(items); i++ {
						if item, ok := items[i].(listItem); ok {
							newQueue = append(newQueue, item.song)
						}
					}
					for i := 0; i < idx; i++ {
						if item, ok := items[i].(listItem); ok {
							newQueue = append(newQueue, item.song)
						}
					}
					m.queue = newQueue
					if m.shuffle {
						r := rand.New(rand.NewSource(time.Now().UnixNano()))
						r.Shuffle(len(m.queue), func(i, j int) {
							m.queue[i], m.queue[j] = m.queue[j], m.queue[i]
						})
					}
					return m.startPlaying(li.song)
				}
			case stateQueue:
				if li, ok := m.list.SelectedItem().(listItem); ok {
					idx := m.list.Index()
					if idx+1 < len(m.queue) {
						m.queue = m.queue[idx+1:]
					} else {
						m.queue = nil
					}
					m.loadQueueList()
					return m.startPlaying(li.song)
				}
			case stateDownload:
				if opt, ok := m.list.SelectedItem().(dlOptionItem); ok {
					m.downloading = true
					m.statusMsg = "Downloading " + opt.ext + "..."
					if m.downloadTarget.AddedAt == "" {
						m.downloadTarget.AddedAt = time.Now().Format(time.RFC3339)
					}
					m.library.Upsert(m.downloadTarget)
					core.SaveLibrary(m.library)

					cmd := doDownloadCmd(m.downloadTarget, opt.args, opt.ext)
					m.state = m.prevState
					m.list = m.prevList
					return m, cmd
				}
			case statePlaylists:
				if i, ok := m.list.SelectedItem().(playlistItem); ok {
					m.selectedPlaylist = i.pl.Name
					m.state = statePlaylistView
					m.loadPlaylistSongsList(i.pl.Name)
				}
			}

		case "f":
			var targetSong models.Song
			isFaving := false

			if m.state == statePlaying {
				m.playing.Favorite = !m.playing.Favorite
				targetSong = m.playing
				isFaving = m.playing.Favorite
			} else if i, ok := m.list.SelectedItem().(listItem); ok {
				targetSong = i.song
				if existing := m.library.FindSong(targetSong.ID); existing != nil {
					existing.Favorite = !existing.Favorite
					targetSong = *existing
					isFaving = existing.Favorite
				} else {
					targetSong.Favorite = true
					isFaving = true
				}
			}

			if targetSong.ID != "" {
				m.library.Upsert(targetSong)
				core.SaveLibrary(m.library)

				if isFaving {
					m.statusMsg = "Added to favorites"
					if targetSong.LocalPath == "" && !m.downloading {
						m.downloading = true
						m.downloadTarget = targetSong
						m.statusMsg = "Added to favorites (Downloading...)"
						return m, doDownloadCmd(targetSong, []string{"-x", "--audio-format", "mp3", "--audio-quality", "0"}, "mp3")
					}
				} else {
					m.statusMsg = "Removed from favorites"
				}
			}

		case "q":
			if m.state != statePlaying && m.state != stateQueue && m.state != stateDownload && m.state != stateCreatePlaylist {
				if i, ok := m.list.SelectedItem().(listItem); ok {
					m.queue = append(m.queue, i.song)
					m.statusMsg = "Added to Queue: " + truncate(i.song.Title, 30)
				}
			}

		case "d":
			if !m.downloading {
				var song models.Song
				if m.state == statePlaying {
					song = m.playing
				} else if i, ok := m.list.SelectedItem().(listItem); ok {
					song = i.song
				}
				if song.ID != "" && song.LocalPath == "" {
					m.downloadTarget = song
					m.prevState = m.state
					m.prevList = m.list
					m.state = stateDownload
					m.loadDownloadMenu()
				} else if song.LocalPath != "" {
					m.statusMsg = "Already downloaded locally"
				}
			}

		case "a":
			var song models.Song
			if m.state == statePlaying {
				song = m.playing
			} else if i, ok := m.list.SelectedItem().(listItem); ok {
				song = i.song
			}
			if song.ID != "" {
				if len(m.library.Playlists) == 0 {
					m.statusMsg = "No playlists — press [4] then [c] to create one"
				} else {
					m.addTargetSong = song
					m.state = stateAddToPlaylist
					m.loadAddToPlaylistList()
				}
			}

		case "c":
			if m.state == statePlaylists {
				m.state = stateCreatePlaylist
				m.createPlaylistInput.Focus()
				return m, textinput.Blink
			}

		case "i":
			if m.state == stateLibrary {
				m.statusMsg = "Scanning ~/Music..."
				return m, tea.Batch(m.spinner.Tick, importLocalFilesCmd(m.library))
			}

		case "U":
			m.statusMsg = "Updating yt-dlp..."
			m.state = stateLoading
			return m, tea.Batch(m.spinner.Tick, updateYTDLPCmd())

		case "J":
			if m.state == stateQueue {
				idx := m.list.Index()
				if idx < len(m.queue)-1 {
					m.queue[idx], m.queue[idx+1] = m.queue[idx+1], m.queue[idx]
					m.loadQueueList()
					m.list.Select(idx + 1)
				}
			} else if m.state == statePlaylistView {
				idx := m.list.Index()
				pl := m.library.FindPlaylist(m.selectedPlaylist)
				if pl != nil && idx < len(pl.SongIDs)-1 {
					pl.SongIDs[idx], pl.SongIDs[idx+1] = pl.SongIDs[idx+1], pl.SongIDs[idx]
					core.SaveLibrary(m.library)
					m.loadPlaylistSongsList(m.selectedPlaylist)
					m.list.Select(idx + 1)
				}
			}

		case "K":
			if m.state == stateQueue {
				idx := m.list.Index()
				if idx > 0 {
					m.queue[idx], m.queue[idx-1] = m.queue[idx-1], m.queue[idx]
					m.loadQueueList()
					m.list.Select(idx - 1)
				}
			} else if m.state == statePlaylistView {
				idx := m.list.Index()
				pl := m.library.FindPlaylist(m.selectedPlaylist)
				if pl != nil && idx > 0 {
					pl.SongIDs[idx], pl.SongIDs[idx-1] = pl.SongIDs[idx-1], pl.SongIDs[idx]
					core.SaveLibrary(m.library)
					m.loadPlaylistSongsList(m.selectedPlaylist)
					m.list.Select(idx - 1)
				}
			}

		case "x":
			if m.state == statePlaylists {
				if i, ok := m.list.SelectedItem().(playlistItem); ok {
					m.library.DeletePlaylist(i.pl.Name)
					core.SaveLibrary(m.library)
					m.statusMsg = "Deleted: " + i.pl.Name
					m.loadPlaylistsList()
				}
			}
			if m.state == statePlaylistView {
				if i, ok := m.list.SelectedItem().(listItem); ok {
					m.library.RemoveSongFromPlaylist(m.selectedPlaylist, i.song.ID)
					core.SaveLibrary(m.library)
					m.statusMsg = "Removed from playlist"
					m.loadPlaylistSongsList(m.selectedPlaylist)
				}
			}
			if m.state == stateQueue {
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.queue) {
					m.queue = append(m.queue[:idx], m.queue[idx+1:]...)
					m.loadQueueList()
					m.statusMsg = "Removed from queue"
				}
			}

		case " ":
			if m.state == statePlaying {
				core.SendMpvToggle(models.MpvSocket)
				m.paused = !m.paused
				core.UpdateNotificationStatus(m.playing.Title, m.playing.Channel, m.paused)
				if m.paused {
					m.statusMsg = "Paused"
				} else {
					m.statusMsg = "Playing"
				}
			}

		case "left":
			if m.state == statePlaying {
				core.SendMpvSeek(models.MpvSocket, -10)
			}
		case "right":
			if m.state == statePlaying {
				core.SendMpvSeek(models.MpvSocket, 10)
			}

		case "+", "=":
			if m.state == statePlaying {
				core.SendMpvVolume(models.MpvSocket, 5)
				m.volume = clamp(m.volume+5, 0, 130)
				m.statusMsg = fmt.Sprintf("Volume: %d%%", m.volume)
			}
		case "-":
			if m.state == statePlaying {
				core.SendMpvVolume(models.MpvSocket, -5)
				m.volume = clamp(m.volume-5, 0, 130)
				m.statusMsg = fmt.Sprintf("Volume: %d%%", m.volume)
			}

		case "[":
			if m.state == statePlaying {
				m.speed = clampFloat(m.speed-0.1, 0.5, 2.0)
				core.SendMpvSpeed(models.MpvSocket, m.speed)
				m.statusMsg = fmt.Sprintf("Speed: %.1fx", m.speed)
			}
		case "]":
			if m.state == statePlaying {
				m.speed = clampFloat(m.speed+0.1, 0.5, 2.0)
				core.SendMpvSpeed(models.MpvSocket, m.speed)
				m.statusMsg = fmt.Sprintf("Speed: %.1fx", m.speed)
			}

		case "n":
			if m.state == statePlaying {
				m.playing.LastPos = m.mpvPos
				m.library.Upsert(m.playing)
				core.SaveLibrary(m.library)
				return m, func() tea.Msg { return songFinishedMsg{} }
			}

		case "s":
			if m.state == statePlaying {
				m.shuffle = !m.shuffle
				if m.shuffle {
					m.statusMsg = "Shuffle: ON"
					if len(m.queue) > 0 {
						r := rand.New(rand.NewSource(time.Now().UnixNano()))
						r.Shuffle(len(m.queue), func(i, j int) {
							m.queue[i], m.queue[j] = m.queue[j], m.queue[i]
						})
					}
				} else {
					m.statusMsg = "Shuffle: OFF"
				}
			}

		case "l":
			if m.state == statePlaying {
				m.repeatMode = (m.repeatMode + 1) % 3
				switch m.repeatMode {
				case repeatOff:
					m.statusMsg = "Repeat: OFF"
				case repeatAll:
					m.statusMsg = "Repeat: ALL"
				case repeatOne:
					m.statusMsg = "Repeat: ONE"
				}
			}

		case "r":
			if m.state == statePlaying {
				m.radioOn = !m.radioOn
				if m.radioOn {
					m.statusMsg = "Artist Radio: ON"
					return m, FetchArtistRadioCmd(m.playing.Channel)
				}
				m.queue = nil
				m.statusMsg = "Artist Radio: OFF"
			}

		case "e":
			if m.state == statePlaying {
				m.eqIndex = (m.eqIndex + 1) % len(core.EqPresets)
				m.statusMsg = "EQ: " + core.EqPresets[m.eqIndex].Name

				return m, func() tea.Msg {
					eqVals := core.EqPresets[m.eqIndex].Values
					af := "silenceremove=start_periods=1:start_threshold=-50dB,loudnorm=I=-14:TP=-1.5:LRA=11,afade=t=in:ss=0:d=1.5"
					if eqVals != "" && eqVals != "0:0:0:0:0:0:0:0:0:0" {
						af += ",@eq:equalizer=" + eqVals
					}
					core.MpvIPC(models.MpvSocket, fmt.Sprintf(`{"command":["set_property","af","%s"]}`, af))
					return nil
				}
			}

		case "t":
			if m.state == statePlaying {
				m.sleepIdx = (m.sleepIdx + 1) % len(sleepOptions)
				mins := sleepOptions[m.sleepIdx]
				if mins == 0 {
					m.sleepEnd = time.Time{}
					m.statusMsg = "Sleep timer: OFF"
				} else {
					m.sleepEnd = time.Now().Add(time.Duration(mins) * time.Minute)
					m.statusMsg = fmt.Sprintf("Sleep timer: %d min", mins)
				}
			}
		}

	case searchResultMsg:
		l := makeList(fmt.Sprintf("Results: \"%s\"", m.query))
		l.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
		l.SetItems(m.toListItems([]models.Song(msg), true))
		m.list = l
		m.state = stateResults

	case trendingMsg:
		l := makeList("Trending Music")
		l.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
		l.SetItems(m.toListItems([]models.Song(msg), true))
		m.list = l
		m.state = stateTrending

	case artistRadioMsg:
		m.queue = append(m.queue, []models.Song(msg)...)
		m.statusMsg = fmt.Sprintf("Radio: %d songs queued", len(m.queue))

	case lyricsReadyMsg:
		m.lyrics = []models.LyricLine(msg)
		if len(m.lyrics) == 0 {
			m.statusMsg = "No lyrics found"
		} else {
			m.playing.Lyrics = m.lyrics
			if existing := m.library.FindSong(m.playing.ID); existing != nil {
				existing.Lyrics = m.lyrics
			}
			m.library.Upsert(m.playing)
			core.SaveLibrary(m.library)
		}

	case ytUpdateDoneMsg:
		if msg.err != nil {
			m.statusMsg = "yt-dlp update failed: " + msg.err.Error()
		} else {
			m.statusMsg = "yt-dlp updated successfully!"
		}
		m.state = stateSearch
		m.textInput.Focus()

	case discoverMsg:
		m.state = stateDiscover
		m.list = makeList("Discover (Recommended for you)")
		m.list.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
		m.list.SetItems(m.toListItems(msg, false))

	case importDoneMsg:
		m.library = msg.lib
		core.SaveLibrary(m.library)
		m.loadLibraryList()
		if msg.added > 0 {
			m.statusMsg = fmt.Sprintf("Imported %d local files", msg.added)
		} else {
			m.statusMsg = "No new local files found"
		}

	case songFinishedMsg:
		if m.repeatMode == repeatOne {
			return m.startPlaying(m.playing)
		}
		if len(m.queue) > 0 {
			if m.repeatMode == repeatAll {
				m.queue = append(m.queue, m.playing)
			}
			next := m.queue[0]
			m.queue = m.queue[1:]
			return m.startPlaying(next)
		}
		if m.repeatMode == repeatAll {
			return m.startPlaying(m.playing)
		}
		if m.radioOn {
			m.statusMsg = "Radio: fetching more..."
			return m, FetchArtistRadioCmd(m.playing.Channel)
		}
		m.state = stateResults
		core.ClearNotification()
		core.ReleaseWakeLock()

	case downloadDoneMsg:
		m.downloading = false
		if msg.path != "" {
			if existing := m.library.FindSong(m.downloadTarget.ID); existing != nil {
				existing.LocalPath = msg.path
				if m.playing.ID == m.downloadTarget.ID {
					m.playing.LocalPath = msg.path
				}
			}
			core.SaveLibrary(m.library)
			m.statusMsg = "Audio download complete"
		} else {
			m.statusMsg = "Video download complete"
		}

	case errMsg:
		m.statusMsg = "Error: " + string(msg)
		m.downloading = false
		if m.state == stateLoading || m.state == stateTrendingLoad {
			m.state = stateSearch
			m.textInput.Focus()
		}
		return m, nil

	case animTickMsg:
		m.animFrame++
		return m, animTickCmd()

	case tickMsg:
		if m.state == statePlaying {
			m.mpvPos = core.GetMpvPos(models.MpvSocket)
			m.paused = core.GetMpvPaused(models.MpvSocket)
			m.updateLyricOffset()

			if m.mpvPos > 1.0 {
				m.trackArmed = true
			}

			if m.trackArmed && core.GetMpvIdle(models.MpvSocket) && !m.paused {
				m.trackArmed = false
				m.playing.LastPos = 0
				m.library.Upsert(m.playing)
				core.SaveLibrary(m.library)
				return m, func() tea.Msg { return songFinishedMsg{} }
			}

			if !m.sleepEnd.IsZero() && time.Now().After(m.sleepEnd) {
				m.sleepEnd = time.Time{}
				m.sleepIdx = 0
				m.queue = nil
				m.trackArmed = false
				core.MpvIPC(models.MpvSocket, `{"command":["stop"]}`)
				m.state = stateResults
				m.statusMsg = "Sleep timer: playback stopped"
				return m, nil
			}
			if !m.scrobbled && m.playing.Duration > 0 &&
				m.mpvPos > 30 &&
				(m.mpvPos > m.playing.Duration/2 || m.mpvPos > 240) {
				m.scrobbled = true
				go core.LastfmScrobble(m.playing, m.library.LastFM)
			}
			return m, tickCmd()
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.state == stateSearch {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	isListState := m.state == stateResults || m.state == stateTrending ||
		m.state == stateLibrary || m.state == statePlaylists ||
		m.state == statePlaylistView || m.state == stateHistory ||
		m.state == stateAddToPlaylist || m.state == stateQueue || m.state == stateDownload || m.state == stateDiscover
	if isListState {
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}
