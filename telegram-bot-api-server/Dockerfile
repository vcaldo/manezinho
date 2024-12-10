FROM ubuntu:24.04 AS builder

# Install dependencies
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y \
    make \
    git \
    zlib1g-dev \
    libssl-dev \
    gperf \
    cmake \
    g++

# Clone and build telegram-bot-api
WORKDIR /build
RUN git clone --recursive https://github.com/tdlib/telegram-bot-api.git && \
    cd telegram-bot-api && \
    mkdir build && \
    cd build && \
    cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=.. .. && \
    cmake --build . --target install

FROM ubuntu:24.04

# Copy only the built binary
COPY --from=builder /build/telegram-bot-api/bin/telegram-bot-api /usr/local/bin/

# Set the binary as entrypoint
ENTRYPOINT ["telegram-bot-api"]