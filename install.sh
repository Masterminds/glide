testRoot() {
	if [ "$(id -u)" != "0" ]; then
		echo "You must be root to run this script"
		exit 1
	fi
}


initArch() {
	ARCH=$(uname -m)
	case $ARCH in
		arm*) ARCH="arm";;
		x86) ARCH="386";;
		x86_64) ARCH="amd64";;
	esac
}

initOS() {
    OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')
}

downloadFile() {
	LATEST_RELEASE_URL="https://api.github.com/repos/Masterminds/glide/releases/latest"
	LATEST_RELEASE_JSON=$(curl -s "$LATEST_RELEASE_URL")
	TAG=$(echo "$LATEST_RELEASE_JSON" | grep 'tag_' | cut -d\" -f4)
	GLIDE_DIST="glide-$TAG-$OS-$ARCH.tar.gz"
	# || true forces this command to not catch error if grep does not find anything
	DOWNLOAD_URL=$(echo "$LATEST_RELEASE_JSON" | grep 'browser_' | cut -d\" -f4 | grep "$GLIDE_DIST") || true
	if [ -z "$DOWNLOAD_URL" ]; then
        echo "Sorry, we dont have a dist for your system: $OS $ARCH"
        echo "You can ask one here: https://github.com/Masterminds/glide/issues"
        exit 1
	else
		GLIDE_TMP_FILE="/tmp/$GLIDE_DIST"
        echo "Downloading $DOWNLOAD_URL"
        curl -L "$DOWNLOAD_URL" -o "$GLIDE_TMP_FILE"
	fi
}

installFile() {
	GLIDE_TMP="/tmp/glide"
	mkdir -p "$GLIDE_TMP"
	tar xf "$GLIDE_TMP_FILE" -C "/tmp/glide"
	GLIDE_TMP_BIN="$GLIDE_TMP/$OS-$ARCH/glide"
	sudo cp "$GLIDE_TMP_BIN" "/usr/local/bin"
}

bye() {
	result=$?
	if [ "$result" != "0" ]; then
		echo "Fail to install glide"
	fi
	exit $result
}

# Execution

#Stop execution on any error
set -e
trap "bye" EXIT
testRoot
initArch
initOS
downloadFile
installFile
# Test if everything ok
GLIDE_VERSION=$(glide -v)
echo "$GLIDE_VERSION installed succesfully"