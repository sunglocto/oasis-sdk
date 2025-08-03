# Oasis SDK for XMPP

> ⚠️ **WARNING**: This project is in early development and APIs are subject
> to change without notice. Use in production at your own peril.

A Go-based wrapper around the Mellium XMPP libraries, designed to simplify 
XMPP implementation in Go projects. This library makes it easier to build 
XMPP-based applications by providing a more straightforward API on top of 
Mellium's comprehensive XMPP stack. It will also be the backend for the 
eventual Oasis Messaging client

## Features

Currently supported XMPP features include:

- **Chat State Notifications**
    - Chatstates
    - Read receipts
    - Delivery Receipts

- **Message Management**
    - Send and receive messages
    - Message reply parsing and sending
    - basic MUC interop

- **HTTP Upload** (XEP-0363)

## Project Structure

A Go-based project developed with Go 1.24.5.

```
├── main.go           # Application entry point
├── types.go          # Type definitions
├── message.go        # Message handling
├── upload.go         # HTTP Upload implementation
├── disco.go          # Service discovery
├── receipts.go       # Message receipt handling
├── chatstates.go     # Chat state management
├── parseFeatures.go  # Feature parsing functionality
└── go.mod            # Go module dependencies
```
## Requirements

- Go 1.24.5 or later

## Usage Example

(Coming soon - The project is in early stage development)

## Want to contribute?

1. Clone the repository
2. Ensure you have Go 1.24.5 or later installed
3. Run the following command:
```bash
# Download dependencies
go mod download
```
## Why This SDK?

While Mellium provides a powerful and complete XMPP implementation, it can be complex to work with directly. This library aims to:

- Simplify common XMPP operations
- Provide sensible defaults
- Abstract away boilerplate code
- Make XMPP integration more accessible to Go developers (such as myself)

## License

This project is licensed under AGPL version 3.0 or later, as found in `./LICENSE`

## Development Status

This project is currently under active development. APIs may change as we continue to improve and expand functionality. Contributions and feedback are welcome!

## Credits

Built on top of the excellent [Mellium XMPP libraries](https://mellium.im/).
