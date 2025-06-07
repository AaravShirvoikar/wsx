# wsx

A simple WebSocket client and server library written in Go that adheres to [RFC 6455](https://datatracker.ietf.org/doc/html/rfc6455) and is fully compliant with the [Autobahn Testsuite](https://github.com/crossbario/autobahn-testsuite).

## Features

- Full WebSocket protocol implementation according to RFC 6455
- Both client and server support
- Proper frame handling with masking/unmasking
- Control frame support (ping/pong, close)
- Message fragmentation support
- UTF-8 validation for text frames
- Fully compliant with the Autobahn Testsuite

## Installation

```bash
go get github.com/AaravShirvoikar/wsx
```

## Usage

See the `examples/` directory for usage examples of both client and server implementations.

## Building and Running Examples

```bash
# Build and run server
make run-server

# Build and run client
make run-client
```
