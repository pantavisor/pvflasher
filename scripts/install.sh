#!/bin/bash
set -euo pipefail

# Colors
Color_Off=''
Red=''
Green=''
Dim=''
Bold_White=''

if [[ -t 1 ]]; then
	Color_Off='\033[0m'
	Red='\033[0;31m'
	Green='\033[0;32m'
	Dim='\033[0;2m'
	Bold_White='\033[1m'
fi

error() {
	echo -e "${Red}error${Color_Off}: $*" >&2
	exit 1
}

info() {
	echo -e "${Dim}$*${Color_Off}"
}

info_bold() {
	echo -e "${Bold_White}$*${Color_Off}"
}

success() {
	echo -e "${Green}$*${Color_Off}"
}

# Configuration
PROJECT_NAME="pvflasher"
BINARY_NAME="pvflasher"
REPO_URL="https://github.com/pantavisor/pvflasher"
API_URL="https://api.github.com/repos/pantavisor/pvflasher"

# Parse arguments
BUILD_FROM_SOURCE=false
VERSION=""

while [[ $# -gt 0 ]]; do
	case $1 in
	--build)
		BUILD_FROM_SOURCE=true
		shift
		;;
	-v | --version)
		VERSION="$2"
		shift 2
		;;
	-h | --help)
		echo "Usage: $0 [OPTIONS] [VERSION]"
		echo ""
		echo "Options:"
		echo "  --build          Build from source instead of downloading"
		echo "  -v, --version    Install specific version"
		echo "  -h, --help       Show this help message"
		echo ""
		echo "Examples:"
		echo "  $0                          # Install latest release"
		echo "  $0 v0.1.0                   # Install specific version"
		echo "  $0 --build                  # Build and install from source"
		echo "  $0 --build v0.1.0           # Build specific version from source"
		exit 0
		;;
	*)
		if [[ -z "$VERSION" && "$1" != -* ]]; then
			VERSION="$1"
		fi
		shift
		;;
	esac
done

# Detect OS
OS=$(uname -s)
case "$OS" in
Linux | Darwin) ;;
*)
	error "Unsupported OS: $OS"
	;;
esac

info "Installing $PROJECT_NAME on $OS..."

# Build from source
if [ "$BUILD_FROM_SOURCE" = true ]; then
	info "Building $PROJECT_NAME from source..."

	# Check for required build tools
	if ! command -v go &>/dev/null; then
		error "Go is required to build from source. Please install Go first."
	fi

	if ! command -v make &>/dev/null; then
		error "Make is required to build from source. Please install Make first."
	fi

	# Check if we're in a git repo with source code
	if [[ -f "Makefile" && -f "go.mod" ]]; then
		info "Building from current directory..."
		if [ -n "$VERSION" ]; then
			info "Checking out version: $VERSION"
			git checkout "$VERSION" 2>/dev/null || error "Failed to checkout version $VERSION"
		fi
	else
		# Clone the repo
		info "Cloning repository..."
		TEMP_DIR=$(mktemp -d)
		trap "rm -rf $TEMP_DIR" EXIT
		git clone "$REPO_URL" "$TEMP_DIR" --depth 1 ${VERSION:+--branch "$VERSION"} || error "Failed to clone repository"
		cd "$TEMP_DIR"
	fi

	# Build
	info "Building binary..."
	make build || error "Build failed"

	# Determine installation directory
	if [ "$OS" = "Darwin" ]; then
		INSTALL_ENV="PVFLASHER_INSTALL"
		INSTALL_DIR="${!INSTALL_ENV:-$HOME/.pvflasher}"
		BIN_DIR="$INSTALL_DIR/bin"
	else
		BIN_DIR="$HOME/.local/bin"
	fi

	EXE="$BIN_DIR/$BINARY_NAME"

	info "Installing to $BIN_DIR..."
	mkdir -p "$BIN_DIR"
	cp "bin/$BINARY_NAME" "$EXE"
	chmod +x "$EXE"

	# Helper to show ~ instead of $HOME
	tildify() {
		if [[ $1 = $HOME/* ]]; then
			echo "${1/$HOME\//\~/}"
		else
			echo "$1"
		fi
	}

	success "$PROJECT_NAME was installed successfully to $(tildify "$EXE")"

	# Setup PATH for Darwin
	if [ "$OS" = "Darwin" ]; then
		if command -v "$BINARY_NAME" >/dev/null; then
			echo ""
			info "Run '$BINARY_NAME --help' to get started"
			exit 0
		fi

		refresh_command=''
		tilde_bin_dir=$(tildify "$BIN_DIR")
		quoted_install_dir=\""${INSTALL_DIR//\"/\\\"}"\"

		if [[ $quoted_install_dir = \"$HOME/* ]]; then
			quoted_install_dir=${quoted_install_dir/$HOME\//\$HOME/}
		fi

		echo ""

		case $(basename "$SHELL") in
		fish)
			commands=(
				"set --export $INSTALL_ENV $quoted_install_dir"
				"set --export PATH \$$INSTALL_ENV/bin \$PATH"
			)
			fish_config="$HOME/.config/fish/config.fish"
			tilde_fish_config=$(tildify "$fish_config")
			mkdir -p "$(dirname "$fish_config")"
			if [[ -w $fish_config ]] || [[ ! -f $fish_config ]]; then
				{
					echo -e '\n# pvflasher'
					for command in "${commands[@]}"; do
						echo "$command"
					done
				} >>"$fish_config"
				info "Added \"$tilde_bin_dir\" to \$PATH in \"$tilde_fish_config\""
				refresh_command="source $tilde_fish_config"
			else
				echo "Manually add the directory to $tilde_fish_config (or similar):"
				for command in "${commands[@]}"; do
					info_bold "  $command"
				done
			fi
			;;
		zsh)
			commands=(
				"export $INSTALL_ENV=$quoted_install_dir"
				"export PATH=\"\$$INSTALL_ENV/bin:\$PATH\""
			)
			zsh_config="$HOME/.zshrc"
			tilde_zsh_config=$(tildify "$zsh_config")
			if [[ -w $zsh_config ]] || [[ ! -f $zsh_config ]]; then
				{
					echo -e '\n# pvflasher'
					for command in "${commands[@]}"; do
						echo "$command"
					done
				} >>"$zsh_config"
				info "Added \"$tilde_bin_dir\" to \$PATH in \"$tilde_zsh_config\""
				refresh_command="exec $SHELL"
			else
				echo "Manually add the directory to $tilde_zsh_config (or similar):"
				for command in "${commands[@]}"; do
					info_bold "  $command"
				done
			fi
			;;
		bash)
			commands=(
				"export $INSTALL_ENV=$quoted_install_dir"
				"export PATH=\"\$$INSTALL_ENV/bin:\$PATH\""
			)
			bash_configs=(
				"$HOME/.bash_profile"
				"$HOME/.bashrc"
				"$HOME/.profile"
			)
			set_manually=true
			for bash_config in "${bash_configs[@]}"; do
				tilde_bash_config=$(tildify "$bash_config")
				if [[ -w $bash_config ]] || [[ ! -f $bash_config ]]; then
					{
						echo -e '\n# pvflasher'
						for command in "${commands[@]}"; do
							echo "$command"
						done
					} >>"$bash_config"
					info "Added \"$tilde_bin_dir\" to \$PATH in \"$tilde_bash_config\""
					refresh_command="source $bash_config"
					set_manually=false
					break
				fi
			done
			if [[ $set_manually = true ]]; then
				echo "Manually add the directory to ~/.bashrc (or similar):"
				for command in "${commands[@]}"; do
					info_bold "  $command"
				done
			fi
			;;
		*)
			echo 'Manually add the directory to ~/.bashrc (or similar):'
			info_bold "  export $INSTALL_ENV=$quoted_install_dir"
			info_bold "  export PATH=\"\$$INSTALL_ENV/bin:\$PATH\""
			;;
		esac

		echo ""
		info "To get started, run:"
		echo ""
		if [[ $refresh_command ]]; then
			info_bold "  $refresh_command"
		fi
		info_bold "  $BINARY_NAME --help"
	else
		echo ""
		success "Installation complete!"
		if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
			echo "Make sure $BIN_DIR is in your PATH to run: $BINARY_NAME"
		else
			info "Run '$BINARY_NAME --help' to get started"
		fi
	fi
	exit 0
fi

# Check for required tools (for download mode)
command -v unzip >/dev/null || error 'unzip is required to install pvflasher'

# Get version if not specified
if [ -z "$VERSION" ]; then
	info "Fetching latest release information..."
	VERSION=$(curl -s "$API_URL/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
fi

if [ -z "$VERSION" ]; then
	error "Could not determine version. Please check $REPO_URL"
fi

info "Installing version: $VERSION"

if [ "$OS" = "Linux" ]; then
	# --- Linux: native package if possible, AppImage fallback ---
	ARCH=$(uname -m)
	case "$ARCH" in
	x86_64)
		FILE_ARCH="x86_64"
		DEB_ARCH="amd64"
		;;
	aarch64 | arm64)
		FILE_ARCH="aarch64"
		DEB_ARCH="arm64"
		;;
	*)
		error "Unsupported architecture: $ARCH"
		;;
	esac

	# Native package filenames omit the 'v' prefix (e.g. 0.0.2, not v0.0.2)
	PKG_VERSION="${VERSION#v}"

	# Detect package manager
	PKG_MGR=""
	if command -v pacman &>/dev/null; then
		PKG_MGR="pacman"
	elif command -v apt &>/dev/null; then
		PKG_MGR="apt"
	elif command -v dnf &>/dev/null; then
		PKG_MGR="dnf"
	elif command -v yum &>/dev/null; then
		PKG_MGR="yum"
	fi

	# Check if we should use package manager (requires sudo/root)
	USE_PKG_MGR=false
	if [ -n "$PKG_MGR" ]; then
		if [ "${EUID:-0}" -eq 0 ] || command -v sudo &>/dev/null; then
			USE_PKG_MGR=true
		fi
	fi

	if [ "$USE_PKG_MGR" = true ] && [ "$PKG_MGR" = "pacman" ]; then
		PKG_FILE="pvflasher-${PKG_VERSION}-1-${FILE_ARCH}.pkg.tar.zst"
		echo "Detected Arch Linux (pacman). Downloading $PKG_FILE..."
		TMP_FILE=$(mktemp /tmp/pvflasher-XXXXXX.pkg.tar.zst)
		curl -fL -o "$TMP_FILE" "$REPO_URL/releases/download/$VERSION/$PKG_FILE"
		echo "Installing (requires sudo)..."
		sudo pacman -U --noconfirm "$TMP_FILE"
		rm -f "$TMP_FILE"

	elif [ "$USE_PKG_MGR" = true ] && [ "$PKG_MGR" = "apt" ]; then
		PKG_FILE="pvflasher_${PKG_VERSION}_${DEB_ARCH}.deb"
		echo "Detected Debian/Ubuntu (apt). Downloading $PKG_FILE..."
		TMP_FILE=$(mktemp /tmp/pvflasher-XXXXXX.deb)
		curl -fL -o "$TMP_FILE" "$REPO_URL/releases/download/$VERSION/$PKG_FILE"
		echo "Installing (requires sudo)..."
		sudo apt install -y "$TMP_FILE"
		rm -f "$TMP_FILE"

	elif [ "$USE_PKG_MGR" = true ] && ([ "$PKG_MGR" = "dnf" ] || [ "$PKG_MGR" = "yum" ]); then
		PKG_FILE="pvflasher-${PKG_VERSION}-1.${FILE_ARCH}.rpm"
		echo "Detected RPM-based distro ($PKG_MGR). Downloading $PKG_FILE..."
		TMP_FILE=$(mktemp /tmp/pvflasher-XXXXXX.rpm)
		curl -fL -o "$TMP_FILE" "$REPO_URL/releases/download/$VERSION/$PKG_FILE"
		echo "Installing (requires sudo)..."
		sudo $PKG_MGR install -y "$TMP_FILE"
		rm -f "$TMP_FILE"

	else
		# Fallback: AppImage works on any Linux without a package manager (or if no sudo)
		if [ -n "$PKG_MGR" ]; then
			echo "Package manager $PKG_MGR found, but sudo is not available. Falling back to local installation..."
		else
			echo "No supported package manager found. Falling back to AppImage..."
		fi

		BIN_DIR="$HOME/.local/bin"
		APP_DIR="$HOME/.local/share/applications"
		ICON_DIR="$HOME/.local/share/icons/hicolor/256x256/apps"
		mkdir -p "$BIN_DIR" "$APP_DIR" "$ICON_DIR"

		APPIMAGE_URL="$REPO_URL/releases/download/$VERSION/PvFlasher-$VERSION-$FILE_ARCH.AppImage"
		echo "Downloading AppImage to $BIN_DIR/$BINARY_NAME..."
		curl -fL -o "$BIN_DIR/$BINARY_NAME.tmp" "$APPIMAGE_URL"
		chmod +x "$BIN_DIR/$BINARY_NAME.tmp"
		mv -f "$BIN_DIR/$BINARY_NAME.tmp" "$BIN_DIR/$BINARY_NAME"

		echo "Downloading icon..."
		curl -sL -o "$ICON_DIR/$BINARY_NAME.png.tmp" "https://raw.githubusercontent.com/pantavisor/pvflasher/main/Icon.png"
		mv -f "$ICON_DIR/$BINARY_NAME.png.tmp" "$ICON_DIR/$BINARY_NAME.png"

		echo "Creating desktop entry..."
		cat >"$APP_DIR/$BINARY_NAME.desktop" <<EOF
[Desktop Entry]
Name=PvFlasher
Comment=Cross-platform USB Image Flasher
Exec=$BIN_DIR/$BINARY_NAME
Icon=$BINARY_NAME
Terminal=false
Type=Application
Categories=System;Utility;
Keywords=usb;flash;image;disk;bmap;
StartupNotify=true
EOF

		echo ""
		success "Installation complete!"
		echo "You can now run '$BINARY_NAME' from your terminal or application menu."
		echo "Note: Make sure $BIN_DIR is in your PATH."
		exit 0
	fi

	echo ""
	success "Installation complete!"
	echo "You can now run '$BINARY_NAME' from your terminal or application menu."

elif [ "$OS" = "Darwin" ]; then
	# --- macOS: zip install to ~/.pvflasher/bin/ ---
	ARCH=$(uname -m)
	case "$ARCH" in
	x86_64) FILE_ARCH="amd64" ;;
	arm64) FILE_ARCH="arm64" ;;
	*)
		error "Unsupported architecture: $ARCH"
		;;
	esac

	INSTALL_ENV="PVFLASHER_INSTALL"
	INSTALL_DIR="${!INSTALL_ENV:-$HOME/.pvflasher}"
	BIN_DIR="$INSTALL_DIR/bin"
	BIN_ENV="\$$INSTALL_ENV/bin"
	EXE="$BIN_DIR/$BINARY_NAME"

	ZIP_URL="$REPO_URL/releases/download/$VERSION/pvflasher-darwin-$FILE_ARCH.zip"
	ZIP_FILE=$(mktemp /tmp/pvflasher-XXXXXX.zip)
	EXTRACT_DIR=$(mktemp -d /tmp/pvflasher-XXXXXX)

	info "Creating installation directory..."
	mkdir -p "$BIN_DIR"

	info "Downloading..."
	curl -fL -o "$ZIP_FILE" "$ZIP_URL" || error "Failed to download from $ZIP_URL"

	info "Extracting..."
	unzip -oqd "$BIN_DIR" "$ZIP_FILE" || error 'Failed to extract'

	# Handle nested directory structure from zip
	if [ -f "$BIN_DIR/pvflasher-darwin-$FILE_ARCH/$BINARY_NAME" ]; then
		mv "$BIN_DIR/pvflasher-darwin-$FILE_ARCH/$BINARY_NAME" "$BIN_DIR/$BINARY_NAME.tmp"
		rm -rf "$BIN_DIR/pvflasher-darwin-$FILE_ARCH"
		mv "$BIN_DIR/$BINARY_NAME.tmp" "$BIN_DIR/$BINARY_NAME"
	fi

	chmod +x "$EXE" || error 'Failed to set permissions'
	rm -f "$ZIP_FILE"

	# Helper to show ~ instead of $HOME
	tildify() {
		if [[ $1 = $HOME/* ]]; then
			echo "${1/$HOME\//\~/}"
		else
			echo "$1"
		fi
	}

	success "$PROJECT_NAME was installed successfully to $(tildify "$EXE")"

	# Check if already in PATH
	if command -v "$BINARY_NAME" >/dev/null; then
		echo ""
		info "Run '$BINARY_NAME --help' to get started"
		exit 0
	fi

	refresh_command=''
	tilde_bin_dir=$(tildify "$BIN_DIR")
	quoted_install_dir=\""${INSTALL_DIR//\"/\\\"}"\"

	if [[ $quoted_install_dir = \"$HOME/* ]]; then
		quoted_install_dir=${quoted_install_dir/$HOME\//\$HOME/}
	fi

	echo ""

	case $(basename "$SHELL") in
	fish)
		commands=(
			"set --export $INSTALL_ENV $quoted_install_dir"
			"set --export PATH $BIN_ENV \$PATH"
		)

		fish_config="$HOME/.config/fish/config.fish"
		tilde_fish_config=$(tildify "$fish_config")

		mkdir -p "$(dirname "$fish_config")"
		if [[ -w $fish_config ]] || [[ ! -f $fish_config ]]; then
			{
				echo -e '\n# pvflasher'
				for command in "${commands[@]}"; do
					echo "$command"
				done
			} >>"$fish_config"

			info "Added \"$tilde_bin_dir\" to \$PATH in \"$tilde_fish_config\""
			refresh_command="source $tilde_fish_config"
		else
			echo "Manually add the directory to $tilde_fish_config (or similar):"
			for command in "${commands[@]}"; do
				info_bold "  $command"
			done
		fi
		;;
	zsh)
		commands=(
			"export $INSTALL_ENV=$quoted_install_dir"
			"export PATH=\"$BIN_ENV:\$PATH\""
		)

		zsh_config="$HOME/.zshrc"
		tilde_zsh_config=$(tildify "$zsh_config")

		if [[ -w $zsh_config ]] || [[ ! -f $zsh_config ]]; then
			{
				echo -e '\n# pvflasher'
				for command in "${commands[@]}"; do
					echo "$command"
				done
			} >>"$zsh_config"

			info "Added \"$tilde_bin_dir\" to \$PATH in \"$tilde_zsh_config\""
			refresh_command="exec $SHELL"
		else
			echo "Manually add the directory to $tilde_zsh_config (or similar):"
			for command in "${commands[@]}"; do
				info_bold "  $command"
			done
		fi
		;;
	bash)
		commands=(
			"export $INSTALL_ENV=$quoted_install_dir"
			"export PATH=\"$BIN_ENV:\$PATH\""
		)

		bash_configs=(
			"$HOME/.bash_profile"
			"$HOME/.bashrc"
			"$HOME/.profile"
		)

		set_manually=true
		for bash_config in "${bash_configs[@]}"; do
			tilde_bash_config=$(tildify "$bash_config")

			if [[ -w $bash_config ]] || [[ ! -f $bash_config ]]; then
				{
					echo -e '\n# pvflasher'
					for command in "${commands[@]}"; do
						echo "$command"
					done
				} >>"$bash_config"

				info "Added \"$tilde_bin_dir\" to \$PATH in \"$tilde_bash_config\""
				refresh_command="source $bash_config"
				set_manually=false
				break
			fi
		done

		if [[ $set_manually = true ]]; then
			echo "Manually add the directory to ~/.bashrc (or similar):"
			for command in "${commands[@]}"; do
				info_bold "  $command"
			done
		fi
		;;
	*)
		echo 'Manually add the directory to ~/.bashrc (or similar):'
		info_bold "  export $INSTALL_ENV=$quoted_install_dir"
		info_bold "  export PATH=\"$BIN_ENV:\$PATH\""
		;;
	esac

	echo ""
	info "To get started, run:"
	echo ""
	if [[ $refresh_command ]]; then
		info_bold "  $refresh_command"
	fi
	info_bold "  $BINARY_NAME --help"
fi
