package ui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rugved-danej/termux-music/internal/core"
	"github.com/rugved-danej/termux-music/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

func StartUI() {
	os.MkdirAll(models.MusicDir, 0755)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--lastfm-auth":
			core.SetupLastFM()
			return
		case "--help", "-h":
			fmt.Println("termux-music — Terminal music player powered by YouTube")
			fmt.Println("")
			fmt.Println("Usage:")
			fmt.Println("  termux-music              Launch the player")
			fmt.Println("  termux-music --lastfm-auth Setup Last.fm scrobbling")
			fmt.Println("")
			fmt.Println("Tabs: [1] Search  [2] Trending  [3] Library  [4] Playlists  [5] History  [6] Queue  [7] Discover")
			fmt.Println("")
			fmt.Println("Now Playing keys:")
			fmt.Println("  Space  Pause/Resume        n    Skip to next song")
			fmt.Println("  ←/→    Seek -/+10s         r    Toggle artist radio")
			fmt.Println("  +/-    Volume               e    Cycle EQ preset")
			fmt.Println("  [ / ]  Speed down/up        s    Toggle shuffle")
			fmt.Println("  l      Cycle repeat mode    t    Sleep timer")
			fmt.Println("  f      Toggle favorite      d    Download MP3")
			fmt.Println("  a      Add to playlist      Esc  Back")
			fmt.Println("")
			return
		}
	}

	exec.Command("pkg", "install", "-y", "socat").Run()
	core.InitMpvDaemon(models.MpvSocket)

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
