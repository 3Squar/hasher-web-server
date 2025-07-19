# GameWebServer

A simple multiplayer game server with real-time communication using WebSockets and FlatBuffers.

## Features

- Real-time multiplayer game server
- WebSocket communication between client and server
- FlatBuffers for efficient data serialization
- GUI client using Gio UI
- Player position tracking and updates

## Project Structure

- `main.go` - Game server with WebSocket handlers
- `client/` - GUI client application
- `schemes/` - FlatBuffer schema definitions
- `generated/` - Auto-generated Go code from FlatBuffer schemas
- `pkg/schema/` - Schema compilation utilities

## Requirements

- Go 1.24.4+

## Running the Server

```bash
go run main.go
```

The server will start and listen for WebSocket connections.

## Running the Client

```bash
cd client
go run main.go
```

This will open a GUI window that connects to the game server.

## Building

```bash
# Build server
go build -o server main.go

# Build client
cd client
go build -o client main.go
```

## Dependencies

- [fasthttp](https://github.com/valyala/fasthttp) - High-performance HTTP server
- [websocket](https://github.com/fasthttp/websocket) - WebSocket implementation
- [FlatBuffers](https://github.com/google/flatbuffers) - Efficient serialization
- [Gio UI](https://gioui.org/) - Cross-platform GUI framework 