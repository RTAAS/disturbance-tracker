##
# Dockerfile for building dtrack binaries
##
FROM docker.io/debian:13

# Update system and TLS certificates
RUN apt-get update && apt-get install -y \
    ca-certificates && \
    update-ca-certificates

# Install build dependencies (go, gui, docs)
RUN apt-get install -y \
    make golang \
    libgl1-mesa-dev xorg-dev libxkbcommon-dev \
    mkdocs

# Basic cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/*
