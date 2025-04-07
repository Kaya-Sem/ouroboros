FROM golang:1.24-bookworm AS base
WORKDIR /build

# Copy the go.mod and go.sum files to the /build directory
COPY go.mod ./

# Install dependencies
RUN go mod download

# Copy the entire source code into the container
COPY . .
RUN go build -o ouroboros
EXPOSE 8080
CMD ["/build/ouroboros"]
