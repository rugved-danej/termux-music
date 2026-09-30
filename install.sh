#!/data/data/com.termux/files/usr/bin/bash

# Termux environment check
if [ -z "$TERMUX_PREFIX" ]; then
	TERMUX_PREFIX="/data/data/com.termux/files/usr"
fi

if [ ! -d "$TERMUX_PREFIX/bin" ]; then
	echo -e "\033[1;31m[!] Error: This script must be run inside Termux.\033[0m"
	exit 1
fi

SCRIPT_NAME="termux-music"
DESTINATION="$TERMUX_PREFIX/bin/$SCRIPT_NAME"

echo -e "\033[36m[*] Installing $SCRIPT_NAME to $DESTINATION...\033[0m"

echo -e "\033[36m[*] Updating system packages and installing dependencies...\033[0m"
pkg update -y >/dev/null 2>&1
pkg install -y mpv python ffmpeg termux-api >/dev/null 2>&1

echo -e "\033[36m[*] Setting up yt-dlp...\033[0m"
pip install --upgrade yt-dlp >/dev/null 2>&1

if [ ! -d "$HOME/storage" ]; then
	echo -e "\033[36m[*] Configuring Android storage permissions...\033[0m"
	termux-setup-storage
	sleep 2
fi

ARCH=$(uname -m)
case "$ARCH" in
	aarch64) GOARCH="arm64" ;;
	armv7l|armv8l|arm) GOARCH="arm" ;;
	x86_64) GOARCH="amd64" ;;
	i686) GOARCH="386" ;;
	*)
		echo -e "\033[1;31m[!] Error: Unsupported architecture $ARCH\033[0m"
		exit 1
		;;
esac

echo -e "\033[36m[*] Downloading pre-compiled binary for $GOARCH...\033[0m"
DOWNLOAD_URL="https://github.com/rugved-danej/termux-music/releases/latest/download/termux-music-$GOARCH"

if ! curl -# -L --fail "$DOWNLOAD_URL" -o "$DESTINATION"; then
	echo -e "\033[1;31m[!] Error: Failed to download the binary. Please ensure there is a release available for your architecture.\033[0m"
	exit 1
fi
chmod +x "$DESTINATION"

echo -e "\033[1;32m[+] Installation successful!\033[0m"
echo -e "\033[1;36mYou can now run the tool anywhere by simply typing: \033[1;32m$SCRIPT_NAME\033[0m"
