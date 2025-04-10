FROM golang:1.24-bookworm AS base
WORKDIR /build

# Copy the go.mod and go.sum files to the /build directory
COPY go.mod go.sum ./

# Install dependencies
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy the entire source code into the container
COPY . .

RUN swag init
RUN go build -o ouroboros

# Explicitly specify UDP ports
EXPOSE 8080/udp
EXPOSE 8081/udp

CMD ["/build/ouroboros"]
