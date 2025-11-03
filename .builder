##
# Dockerfile used to generate binaries from Julia Language source code.
##
FROM docker.io/debian:13

# Update system and TLS certificates
RUN apt-get update && apt-get install -y \
    ca-certificates && \
    update-ca-certificates

# Install build dependencies (go, gui)
RUN apt-get install -y \
    make golang \
    gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev

# Cache golang dependencies
# TODO: 'go get fyne' requires --tty, but not available in 'go build'
#RUN echo "module temp" >go.mod && \
#    go get fyne.io/fyne/v2 && \
#    go mod tidy && \
#    rm -f go.mod go.sum

# Basic cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/*
