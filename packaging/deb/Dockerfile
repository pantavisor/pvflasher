FROM debian:bookworm

RUN apt-get update && apt-get install -y \
    dpkg \
    dpkg-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /work
