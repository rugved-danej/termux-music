package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/rugved-danej/termux-music/internal/models"
)

var EqPresets = []struct {
	Name   string
	Values string
}{
	{"Flat", "0:0:0:0:0:0:0:0:0:0"},
	{"Bass Boost", "6:5:4:2:0:0:0:0:0:0"},
	{"Treble", "0:0:0:0:0:2:4:5:6:6"},
	{"Pop", "1:0:1:2:3:2:0:0:1:2"},
	{"Rock", "4:3:2:1:0:1:2:3:3:2"},
	{"Classical", "0:0:0:0:0:0:0:0:3:4"},
	{"Deep", "5:4:3:2:0:0:-2:-3:-3:-3"},
}

func InitMpvDaemon(socket string) {
	exec.Command("killall", "mpv").Run()
	os.Remove(socket)
	cmd := exec.Command("mpv", "--idle", "--really-quiet", "--no-video", fmt.Sprintf("--input-ipc-server=%s", socket))
	cmd.Start()
}

func MpvIPC(socket, jsonCmd string) map[string]interface{} {
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil
	}
	defer conn.Close()
	fmt.Fprintf(conn, "%s\n", jsonCmd)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, `"error":`) {
			var res map[string]interface{}
			json.Unmarshal([]byte(text), &res)
			return res
		}
	}
	return nil
}

func GetMpvPos(socket string) float64 {
	res := MpvIPC(socket, `{"command":["get_property","time-pos"]}`)
	if res == nil {
		return 0
	}
	if v, ok := res["data"].(float64); ok {
		return v
	}
	return 0
}

func GetMpvPaused(socket string) bool {
	res := MpvIPC(socket, `{"command":["get_property","pause"]}`)
	if res == nil {
		return false
	}
	if v, ok := res["data"].(bool); ok {
		return v
	}
	return false
}

func getMpvVolume(socket string) int {
	res := MpvIPC(socket, `{"command":["get_property","volume"]}`)
	if res == nil {
		return 100
	}
	if v, ok := res["data"].(float64); ok {
		return int(v)
	}
	return 100
}

func GetMpvIdle(socket string) bool {
	res := MpvIPC(socket, `{"command":["get_property","idle-active"]}`)
	if res == nil {
		return true
	}
	if v, ok := res["data"].(bool); ok {
		return v
	}
	return true
}

func SendMpvToggle(socket string) {
	MpvIPC(socket, `{"command":["cycle","pause"]}`)
}

func SendMpvSeek(socket string, secs int) {
	MpvIPC(socket, fmt.Sprintf(`{"command":["seek",%d,"relative"]}`, secs))
}

func SendMpvVolume(socket string, delta int) {
	MpvIPC(socket, fmt.Sprintf(`{"command":["add","volume",%d]}`, delta))
}

func SendMpvSpeed(socket string, speed float64) {
	MpvIPC(socket, fmt.Sprintf(`{"command":["set_property","speed",%f]}`, speed))
}

func AndroidNotification(title, channel, status string) {
	t := title
	if len([]rune(t)) > 45 {
		runes := []rune(t)
		t = string(runes[:42]) + "..."
	}

	toggleCmd := `echo '{"command":["cycle","pause"]}' | socat - ` + models.MpvSocket
	seekFwdCmd := `echo '{"command":["seek",30,"relative"]}' | socat - ` + models.MpvSocket
	seekBwdCmd := `echo '{"command":["seek",-30,"relative"]}' | socat - ` + models.MpvSocket
	nextCmd := `echo '{"command":["stop"]}' | socat - ` + models.MpvSocket

	exec.Command("termux-notification",
		"--id", "tmusicplayer",
		"--title", t,
		"--content", channel+" | "+status,
		"--icon", "music_note",
		"--priority", "high",
		"--ongoing",
		"--type", "media",
		"--media-previous", seekBwdCmd,
		"--media-pause", toggleCmd,
		"--media-play", toggleCmd,
		"--media-next", nextCmd,
		"--button1", "-30s",
		"--button1-action", seekBwdCmd,
		"--button2", "Pause/Play",
		"--button2-action", toggleCmd,
		"--button3", "+30s | Skip",
		"--button3-action", seekFwdCmd+"||"+nextCmd,
	).Run()
}

func UpdateNotificationStatus(title, channel string, paused bool) {
	status := "Playing"
	if paused {
		status = "Paused"
	}
	AndroidNotification(title, channel, status)
}

func ClearNotification() {
	exec.Command("termux-notification-remove", "tmusicplayer").Run()
}

func AcquireWakeLock() {
	exec.Command("termux-wake-lock").Run()
}

func ReleaseWakeLock() {
	exec.Command("termux-wake-unlock").Run()
}
