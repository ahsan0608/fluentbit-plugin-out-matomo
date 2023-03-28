# Use Debian as our base image
FROM debian:buster-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    build-essential \
    gnupg2 && \
    rm -rf /var/lib/apt/lists/*

RUN echo "deb http://deb.debian.org/debian buster-backports main" >> /etc/apt/sources.list.d/backports.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends -t buster-backports golang-go

# Create a directory for our Go module and set it as the working directory
WORKDIR /go/src/app
COPY matomo.go .

# Initialize a new Go module and install the Fluent Bit SDK
RUN go mod init app && \
    go get github.com/fluent/fluent-bit-go/output && \
    go mod tidy

# Build the plugin using the Fluent Bit Go SDK
RUN go build -buildmode=c-shared -o out_matomo.so matomo.go

# Use the official Fluent Bit image as our base
FROM fluent/fluent-bit:1.8.8-debug

# Copy the plugin into the Fluent Bit image
COPY --from=0 /go/src/app/out_matomo.so /fluent-bit/bin/out_matomo.so

COPY plugins.conf /fluent-bit/etc/plugins.conf

# Copy the configuration file into the Fluent Bit image
COPY fluent-bit.conf /fluent-bit/etc/fluent-bit.conf

# Set the Fluent Bit configuration environment variable
ENV FLUENT_BIT_CONF=/fluent-bit/etc/fluent-bit.conf