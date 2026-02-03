FROM ubuntu:20.04

# Set non-interactive to avoid prompts during build
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
	# Build essentials
	build-essential \
	git \
	wget \
	curl \
	pkg-config \
	ca-certificates \
	# Image processing tools
	imagemagick \
	file \
	# Fyne GUI dependencies
	libgl1-mesa-dev \
	xorg-dev \
	libxcursor-dev \
	libxrandr-dev \
	libxi-dev \
	libxinerama-dev \
	libxkbcommon-dev \
	libxxf86vm-dev \
	libc6 \
	libc6-dev \
	libssl-dev \
	# Font rendering dependencies
	libfontconfig1 \
	libfreetype6 \
	libharfbuzz0b \
	&& rm -rf /var/lib/apt/lists/*

# Download and install linuxdeploy
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then \
        LINUXDEPLOY_URL="https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage"; \
        APPIMAGETOOL_URL="https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage"; \
        GO_URL="https://go.dev/dl/go1.24.12.linux-amd64.tar.gz"; \
    elif [ "$ARCH" = "aarch64" ]; then \
        LINUXDEPLOY_URL="https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-aarch64.AppImage"; \
        APPIMAGETOOL_URL="https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-aarch64.AppImage"; \
        GO_URL="https://go.dev/dl/go1.24.12.linux-arm64.tar.gz"; \
    else \
        echo "Unsupported architecture: $ARCH" && exit 1; \
    fi && \
    wget "$LINUXDEPLOY_URL" -O /usr/local/bin/linuxdeploy && \
    chmod +x /usr/local/bin/linuxdeploy && \
    wget "$APPIMAGETOOL_URL" -O /usr/local/bin/appimagetool && \
    chmod +x /usr/local/bin/appimagetool && \
    wget "$GO_URL" -O /tmp/go.tar.gz && \
    tar -xzf /tmp/go.tar.gz -C /usr/local && \
    rm /tmp/go.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/root/go
ENV PATH=$PATH:/root/go/bin

WORKDIR /work
