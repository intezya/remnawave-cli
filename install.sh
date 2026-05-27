#!/usr/bin/env sh
set -eu

REPO="${REPO:-intezya/remnawave-cli}"
BIN_NAME="${BIN_NAME:-remnawave-cli}"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-}"

need() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "error: required command not found: $1" >&2
		exit 1
	fi
}

detect_os() {
	case "$(uname -s)" in
		Linux) echo "linux" ;;
		Darwin) echo "darwin" ;;
		*) echo "error: unsupported OS: $(uname -s)" >&2; exit 1 ;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64 | amd64) echo "amd64" ;;
		arm64 | aarch64) echo "arm64" ;;
		*) echo "error: unsupported architecture: $(uname -m)" >&2; exit 1 ;;
	esac
}

choose_install_dir() {
	if [ -n "$INSTALL_DIR" ]; then
		echo "$INSTALL_DIR"
		return
	fi

	if [ -d /usr/local/bin ] && { [ -w /usr/local/bin ] || command -v sudo >/dev/null 2>&1; }; then
		echo "/usr/local/bin"
		return
	fi

	echo "$HOME/.local/bin"
}

install_binary() {
	src="$1"
	dst_dir="$2"
	dst="$dst_dir/$BIN_NAME"

	if [ ! -d "$dst_dir" ]; then
		mkdir -p "$dst_dir"
	fi

	if [ -w "$dst_dir" ]; then
		install -m 0755 "$src" "$dst"
	elif command -v sudo >/dev/null 2>&1; then
		sudo install -m 0755 "$src" "$dst"
	else
		echo "error: $dst_dir is not writable and sudo is unavailable" >&2
		exit 1
	fi
}

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{print $1}'
	else
		return 1
	fi
}

verify_checksum() {
	archive="$1"
	checksums="$2"

	expected="$(grep "[[:space:]]$(basename "$archive")$" "$checksums" | awk '{print $1}' || true)"
	if [ -z "$expected" ]; then
		echo "warning: checksum for $(basename "$archive") not found, skipping verification" >&2
		return
	fi

	if ! actual="$(sha256 "$archive")"; then
		echo "warning: sha256 tool not found, skipping checksum verification" >&2
		return
	fi

	if [ "$actual" != "$expected" ]; then
		echo "error: checksum mismatch for $(basename "$archive")" >&2
		exit 1
	fi
}

need curl
need tar

os="$(detect_os)"
arch="$(detect_arch)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

if [ "$VERSION" = "latest" ]; then
	api_url="https://api.github.com/repos/$REPO/releases/latest"
else
	api_url="https://api.github.com/repos/$REPO/releases/tags/$VERSION"
fi

release_json="$tmp_dir/release.json"
curl -fsSL "$api_url" -o "$release_json" || {
	echo "error: unable to fetch release metadata from $api_url" >&2
	exit 1
}

asset_url="$(
	sed -n 's/.*"browser_download_url": *"\([^"]*\)".*/\1/p' "$release_json" |
		grep -- "-${os}-${arch}\\.tar\\.gz$" |
		head -n 1
)"

if [ -z "$asset_url" ]; then
	echo "error: no release asset found for ${os}/${arch}" >&2
	exit 1
fi

archive="$tmp_dir/$(basename "$asset_url")"
curl -fL "$asset_url" -o "$archive"

checksums_url="$(
	sed -n 's/.*"browser_download_url": *"\([^"]*\)".*/\1/p' "$release_json" |
		grep -- "/checksums\\.txt$" |
		head -n 1 || true
)"

if [ -n "$checksums_url" ]; then
	curl -fsSL "$checksums_url" -o "$tmp_dir/checksums.txt"
	verify_checksum "$archive" "$tmp_dir/checksums.txt"
fi

extract_dir="$tmp_dir/extract"
mkdir -p "$extract_dir"
tar -xzf "$archive" -C "$extract_dir"
binary="$(find "$extract_dir" -type f -name "${BIN_NAME}*" | head -n 1)"
if [ -z "$binary" ]; then
	echo "error: $BIN_NAME binary not found in release archive" >&2
	exit 1
fi

install_dir="$(choose_install_dir)"
install_binary "$binary" "$install_dir"

case ":$PATH:" in
	*":$install_dir:"*) ;;
	*) echo "warning: $install_dir is not in PATH" >&2 ;;
esac

echo "$BIN_NAME installed to $install_dir/$BIN_NAME"
"$install_dir/$BIN_NAME" --version
