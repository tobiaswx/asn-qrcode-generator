# QR Code Label Generator

A Go application for generating PDF sheets of QR code labels, designed for Avery L4731REV-25 label sheets. The tool can be used both as a CLI application and as a HTTP server.

## Features

- Generates QR codes with customizable prefixes and numbering
- Supports Avery L4731REV-25 label sheets (189 labels per page, 7×27 grid)
- Can be run as a CLI tool or HTTP server
- Configurable number format with leading zeros
- Debug mode with visible label borders
- Temporary file cleanup
- PDF output with precise label positioning

## Installation

```bash
go get github.com/tobiaswx/asn-qrcode-generator
```

## Usage

### CLI Mode

Basic usage with default settings:
```bash
go run main.go
```

Generate multiple pages with custom settings:
```bash
go run main.go -start 1000 -prefix "ASN" -pages 2 -zeros 5 -output "labels.pdf"
```

Available flags:
- `-start`: Starting number (default: 1)
- `-prefix`: Prefix for numbers (default: "ASN")
- `-pages`: Number of pages to generate (default: 1)
- `-output`: Output PDF filename (default: "labels.pdf")
- `-borders`: Show label borders for debugging (default: false)
- `-zeros`: Number of leading zeros in the number (default: 4)

### Server Mode

Start the HTTP server:
```bash
go run main.go -serve -port 8080
```

Generate labels via HTTP request:
```
http://localhost:8080/generate?start=1000&prefix=ASN&pages=2&zeros=5&borders=true
```

Query parameters:
- `start`: Starting number
- `prefix`: Prefix for numbers
- `pages`: Number of pages
- `zeros`: Number of leading zeros
- `borders`: Show borders (true/false)

## Label Sheet Specifications

The generator is configured for Avery L4731REV-25 label sheets with the following specifications:
- 189 labels per page (7×27 grid)
- Label dimensions: 25.4mm × 10.0mm
- Horizontal gutter: 2.55mm
- Left margin: 8.45mm
- Top margin: 13.5mm
- QR code size: 9.0mm

## Docker

The application is available as a Docker image from GitHub Container Registry:

```bash
docker pull ghcr.io/tobiaswx/asn-qrcode-generator:latest
```

Run the container:
```bash
docker run -p 8080:8080 ghcr.io/tobiaswx/asn-qrcode-generator:latest
```

Available platforms:
- linux/amd64
- linux/arm64

## Dependencies

- github.com/boombuler/barcode
- github.com/go-pdf/fpdf

## Building from Source

```bash
git clone https://github.com/tobiaswx/asn-qrcode-generator
cd asn-qrcode-generator
go build
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)