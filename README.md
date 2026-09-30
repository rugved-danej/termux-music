# Termux Music

A powerful, fast, and gorgeous terminal music player designed specifically for Termux on Android. Powered by YouTube, rendered via Bubble Tea, and built with Go.

## Features

* **Native MPV Integration:** True gapless playback via a background daemon socket
* **Multi-Format Downloads:** Download offline tracks directly to your phone via `yt-dlp` (MP3, M4A, FLAC)
* **Real-Time Lyrics:** Synchronized scrolling lyrics powered by the LRCLib API
* **Advanced Animations:** Hardware-accelerated 16FPS custom progress bars and bouncing equalizers
* **Library Management:** Local storage syncing, custom playlists, and offline caching
* **Algorithmic Discovery:** Personalized "Discover" tab based on your listening history
* **Last.fm Scrobbling:** Built-in support for live scrobbling and "Now Playing" updates

## Installation

### Prerequisites
To use background play, wake-locks, and media notifications, you **must** install the official Termux:API app.
* [Download Termux:API (v0.53.0) APK](https://github.com/termux/termux-api/releases/download/v0.53.0/termux-api-app_v0.53.0+github.debug.apk)

### Install Script
Install Termux Music instantly with a single command. This script will automatically set up Android storage permissions, install `mpv`/`yt-dlp`/`termux-api`, and download the pre-compiled binary for your architecture. Go is **not** required.

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/rugved-danej/termux-music/main/install.sh)"
```

## Usage

Start the player from anywhere in your terminal:
```bash
termux-music
```
Upon the first launch, it will instantly load. Use the **Tabs** at the top (`[1]`, `[2]`, `[3]`, etc.) to navigate between Search, Trending, and your Library. 

### Data & Storage
All your downloaded songs, cached thumbnails, and playlists are saved directly to your phone's internal storage, so you can access them with any other Android music player if you want!
* **Downloads:** `Internal Storage/Music/termux-music/`
* **Database:** `Internal Storage/Music/termux-music/library.json`

## Uninstallation

If you ever want to completely remove the app, run:
```bash
rm $PREFIX/bin/termux-music
```
*(Note: This purely deletes the executable. Your downloaded songs and library database will safely remain in your Android `Music` folder unless you delete them yourself!)*

## Keybindings

### Global Navigation
* `1` - Search
* `2` - Trending
* `3` - Library
* `4` - Playlists
* `5` - History
* `6` - Queue
* `7` - Discover (Personalized Recommendations)
* `Esc` - Toggle between Search and the "Now Playing" screen

### Lists & Queue Management
* `Enter` - Play song immediately (clears current track)
* `q` - Add song to Up Next Queue (without interrupting playback)
* `d` - Download song locally
* `f` - Toggle Favorite
* `a` - Add to a Playlist
* `J / K` - Move items up/down in lists (Queue, Playlists)
* `x` - Remove item from current list
* `i` - Import local files (Library tab only)
* `U` - Force update `yt-dlp` to the latest version
* `/` - Filter/Search the current view

### Now Playing Controls
* `Space` - Play / Pause
* `Left / Right` - Seek backward / forward (10s)
* `+ / -` - Volume up / down
* `[ / ]` - Playback speed decrease / increase
* `n` - Skip to next track
* `s` - Toggle shuffle mode
* `l` - Toggle loop mode (All / One / Off)
* `r` - Toggle infinite Artist Radio
* `e` - Cycle visual equalizer presets
* `t` - Set sleep timer (15m, 30m, 1h, 2h)

## License
MIT License - Copyright (c) 2026 Rugved
