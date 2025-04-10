# Ouroboros

A distributed peer-to-peer local-network system built with Go, supporting UDP-based peer discovery and communication.

## Features

- UDP-based peer discovery
- HTTP API for monitoring and management
- Docker support for easy deployment
- Multi-instance support with host networking

## Architecture

The system consists of:
- UDP server for peer discovery and messaging
- HTTP API server for monitoring and management
- Peer management system

## Port Configuration

Each instance uses:
- UDP_PORT: Port for receiving UDP messages (default: 8080)
- HTTP_PORT: Port for HTTP API (default: 8082)
- BROADCAST_PORT: Port for sending broadcast messages (default: 8080)

## Development

### Building

```bash
# Build the Docker image
docker build -t ouroboros .
```

### Running a Single Instance

```bash
# Run with default configuration
docker run --rm --network=host \
  -e NAME=my-client \
  -e UDP_PORT=8080 \
  -e HTTP_PORT=8082 \
  -e BROADCAST_PORT=8080 \
  ouroboros
```

### Running Multiple Instances

Use docker-compose to run multiple instances:

```bash
# Start two instances
docker-compose up --build
```

This will start:
- Instance 1: UDP 8080, HTTP 8082
- Instance 2: UDP 8081, HTTP 8083

Both instances will broadcast to port 8080.

## API Documentation

The HTTP API is documented using Swagger and can be accessed at:
- Instance 1: `http://localhost:8082/docs/`
- Instance 2: `http://localhost:8083/docs/`

### Available Endpoints

- `GET /peers`: List all discovered peers
  ```json
  [
    {
      "name": "peer-name",
      "address": "ip:port"
    }
  ]
  ```

## Environment Variables

- `NAME`: Client name (default: "default")
- `UDP_PORT`: Port for receiving UDP messages
- `HTTP_PORT`: Port for HTTP API
- `BROADCAST_PORT`: Port for sending broadcast messages

## Network Configuration

The system uses host networking mode to ensure proper UDP broadcast functionality. This means:
- Containers share the host's network stack
- UDP broadcasts work across the network
- Each instance needs unique ports

## Troubleshooting

1. If instances can't discover each other:
   - Ensure they're using different UDP ports
   - Check that BROADCAST_PORT is set correctly
   - Verify host networking is working

2. If HTTP API is not accessible:
   - Check that HTTP_PORT is not in use
   - Verify the port is properly exposed
