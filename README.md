# QR Code Label Generator for Paperless-ngx

A Go application for generating PDF sheets of QR code labels for [Paperless-ngx](https://github.com/paperless-ngx/paperless-ngx) ASN (Archive Serial Number) labels. Designed for Avery L4731REV-25 and compatible label sheets (see Label Sheet Specifications for a complete list), this tool helps you create physical labels for your document management system. It can be used both as a CLI application and as a HTTP server.

## Features

- Generates QR codes for Paperless-ngx ASNs with customizable prefixes
- Perfect for creating physical labels for your paper documents before scanning
- Supports multiple label sheet formats including Avery L4731REV-25 and HERMA 10001 (189 labels per page, 7×27 grid)
- Can be run as a CLI tool or HTTP server
- Configurable number format with leading zeros
- Debug mode with visible label borders
- Temporary file cleanup
- PDF output with precise label positioning

## Public Instance

A public instance is available at https://asn-qrcode-generator.tobiaswx.com for quick label generation without the need to set up your own server. This instance is hosted on a private server and tunneled via Cloudflare, so it should be available most of the time (barring power or internet outages).

Example usage:
```
https://asn-qrcode-generator.tobiaswx.com/generate?pages=25
```
This will generate 25 pages of labels, starting with ASN0001 up to ASN4725 (189 labels per page). The example uses the default values (prefix="ASN", start=1, zeros=4) for simplicity - all parameters can be customized as needed. PDF generation for that example typically takes 5-10 seconds.

For immediate access to the most commonly used configuration (25 pages with default settings), you can download a pre-generated PDF directly:
```
https://hel1.your-objectstorage.com/tobiaswx-github-public/asn-qrcode-generator/asn0001-asn4725_25pages_default.pdf
```

Note: While the public instance is provided as a convenience, for production use or if you need guaranteed availability, consider running your own instance using the instructions below.

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

Compatible label sheets:
- Avery L4731REV-25 (primary supported format)
- HERMA 10001 (confirmed compatible due to identical dimensions)

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