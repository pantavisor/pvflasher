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

# Install Go 1.24
RUN wget https://go.dev/dl/go1.24.12.linux-amd64.tar.gz -O /tmp/go.tar.gz && \
	tar -xzf /tmp/go.tar.gz -C /usr/local && \
	rm /tmp/go.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/root/go
ENV PATH=$PATH:/root/go/bin

# Download and install linuxdeploy
RUN wget https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage -O /usr/local/bin/linuxdeploy && \
	chmod +x /usr/local/bin/linuxdeploy

# Download and install appimagetool
RUN wget https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage -O /usr/local/bin/appimagetool && \
	chmod +x /usr/local/bin/appimagetool

WORKDIR /work
