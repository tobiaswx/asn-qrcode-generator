package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/go-pdf/fpdf"
)

const (
	// Label sheet constants (Avery L4731REV-25)
	labelsPerPage = 189
	labelsAcross  = 7
	labelsDown    = 27

	// Label dimensions in millimeters
	labelWidth    = 25.4
	labelHeight   = 10.0
	labelGutterX  = 2.55
	marginLeft    = 8.45
	marginTop     = 13.5
	qrCodeSize    = 9.0
	qrCodeMarginX = 0.5
	qrCodeOffsetY = 0.5
)

type config struct {
	startNumber  int
	prefix       string
	pages        int
	outputFile   string
	showBorders  bool
	leadingZeros int
}

// tempFiles keeps track of temporary files we need to clean up
type tempFiles struct {
	files []string
	mu    sync.Mutex
}

func (tf *tempFiles) add(filename string) {
	tf.mu.Lock()
	tf.files = append(tf.files, filename)
	tf.mu.Unlock()
}

func (tf *tempFiles) cleanup() {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	for _, f := range tf.files {
		os.Remove(f)
	}
	tf.files = nil
}

func main() {
	serveFlag := flag.Bool("serve", false, "Run as HTTP server")
	port := flag.String("port", "8080", "HTTP server port")

	cfg := parseFlags()

	if *serveFlag {
		startServer(*port)
	} else {
		if err := generatePDF(cfg); err != nil {
			log.Fatalf("Error generating PDF: %v", err)
		}
		fmt.Printf("Generated PDF file: %s\n", cfg.outputFile)
	}
}

func parseFlags() config {
	cfg := config{}

	flag.IntVar(&cfg.startNumber, "start", 1, "Starting ASN number")
	flag.StringVar(&cfg.prefix, "prefix", "ASN", "Prefix for ASN numbers")
	flag.IntVar(&cfg.pages, "pages", 1, "Number of pages to generate")
	flag.StringVar(&cfg.outputFile, "output", "labels.pdf", "Output PDF file")
	flag.BoolVar(&cfg.showBorders, "borders", false, "Show label borders (for debugging)")
	flag.IntVar(&cfg.leadingZeros, "zeros", 4, "Number of leading zeros in the number")

	flag.Parse()

	// Ensure output directory exists
	dir := filepath.Dir(cfg.outputFile)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Could not create output directory: %v", err)
		}
	}

	return cfg
}

func startServer(port string) {
	http.HandleFunc("/generate", handleGenerate)
	http.HandleFunc("/", handleRoot)
	log.Printf("Starting server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	hostname, _ := os.Hostname()
	info := fmt.Sprintf(`
    _    ____  _   _     ___  ____      ____          _      
   / \  / ___|| \ | |   / _ \|  _ \    / ___|___   __| | ___ 
  / _ \ \___ \|  \| |  | | | | |_) |  | |   / _ \ / _  |/ _ \
 / ___ \ ___) | |\  |  | |_| |  _ <   | |__| (_) | (_| |  __/
/_/   \_\____/|_| \_|   \__\_\_| \_\   \____\___/ \__,_|\___|
                                             Label Generator %s

Server Information:
------------------
Hostname: %s
Time: %s
Version: v1.0.0

API Usage:
----------
Generate labels: GET /generate
Parameters:
  - start    : Starting ASN number (default: 1)
  - prefix   : Prefix for ASN (default: "ASN")
  - pages    : Number of pages (default: 1)
  - zeros    : Number of leading zeros (default: 4)
  - borders  : Show borders, true/false (default: false)

Examples:
--------
Basic usage:
  /generate?start=1000&prefix=ASN&pages=1

With all parameters:
  /generate?start=1000&prefix=ASN&pages=2&zeros=5&borders=true

Label Sheet Info:
---------------
Type: Avery L4731REV-25
Layout: 7 x 27 (189 labels per page)
Size: 25.4mm x 10.0mm

For more information visit:
https://github.com/tobiaswx/asn-qrcode-generator
`, os.Args[0], hostname, time.Now().Format(time.RFC1123))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, info)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters with defaults
	startNumber, _ := strconv.Atoi(r.URL.Query().Get("start"))
	pages, _ := strconv.Atoi(r.URL.Query().Get("pages"))
	leadingZeros, _ := strconv.Atoi(r.URL.Query().Get("zeros"))
	showBorders, _ := strconv.ParseBool(r.URL.Query().Get("borders"))

	// Set defaults if not provided
	if startNumber == 0 {
		startNumber = 1
	}
	if pages == 0 {
		pages = 1
	}
	if leadingZeros == 0 {
		leadingZeros = 4
	}

	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		prefix = "ASN"
	}

	// Convert to config
	cfg := config{
		startNumber:  startNumber,
		prefix:       prefix,
		pages:        pages,
		outputFile:   fmt.Sprintf("asn-%d.pdf", startNumber),
		showBorders:  showBorders,
		leadingZeros: leadingZeros,
	}

	// Generate PDF
	if err := generatePDF(cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send file
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", cfg.outputFile))
	http.ServeFile(w, r, cfg.outputFile)

	// Clean up the file after sending
	defer os.Remove(cfg.outputFile)
}

func generatePDF(cfg config) error {
	tf := &tempFiles{
		files: make([]string, 0, labelsPerPage),
	}
	// Ensure cleanup happens after we're done
	defer tf.cleanup()

	// Create PDF
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginLeft, marginTop, marginLeft)

	// Generate labels for requested number of pages
	for page := 0; page < cfg.pages; page++ {
		pdf.AddPage()

		startNum := cfg.startNumber + (page * labelsPerPage)
		if err := generatePage(pdf, startNum, cfg, tf); err != nil {
			return fmt.Errorf("error generating page %d: %v", page+1, err)
		}
	}

	if err := pdf.OutputFileAndClose(cfg.outputFile); err != nil {
		return fmt.Errorf("error saving PDF: %v", err)
	}

	return nil
}

func generatePage(pdf *fpdf.Fpdf, startNum int, cfg config, tf *tempFiles) error {
	for row := 0; row < labelsDown; row++ {
		for col := 0; col < labelsAcross; col++ {
			currentNum := startNum + (row * labelsAcross) + col

			// Calculate position
			x := marginLeft + float64(col)*(labelWidth+labelGutterX)
			y := marginTop + float64(row)*labelHeight

			// Generate QR code
			text := fmt.Sprintf("%s%0*d", cfg.prefix, cfg.leadingZeros, currentNum)
			qrPath, err := generateQR(text, tf)
			if err != nil {
				return fmt.Errorf("error generating QR code for %s: %v", text, err)
			}

			// Add QR code to PDF
			pdf.Image(qrPath, x, y+qrCodeOffsetY, qrCodeSize, qrCodeSize, false, "", 0, "")

			// Add text
			pdf.SetFont("Helvetica", "", 8)
			pdf.Text(x+qrCodeSize+qrCodeMarginX, y+labelHeight/2, text)

			// Draw border if enabled
			if cfg.showBorders {
				pdf.Rect(x, y, labelWidth, labelHeight, "D")
			}
		}
	}
	return nil
}

func generateQR(text string, tf *tempFiles) (string, error) {
	// Generate QR code
	qrCode, err := qr.Encode(text, qr.M, qr.Auto)
	if err != nil {
		return "", fmt.Errorf("failed to encode QR code: %v", err)
	}

	// Scale QR code to required size
	qrCode, err = barcode.Scale(qrCode, 100, 100)
	if err != nil {
		return "", fmt.Errorf("failed to scale QR code: %v", err)
	}

	// Convert to RGBA
	bounds := qrCode.Bounds()
	rgbaImg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgbaImg.Set(x, y, qrCode.At(x, y))
		}
	}

	// Create temporary file for QR code
	tmpFile, err := os.CreateTemp("", "asn-label-*.png")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}

	// Add the file to our tracking list
	tf.add(tmpFile.Name())

	// Save QR code to temp file
	if err := png.Encode(tmpFile, rgbaImg); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to encode PNG: %v", err)
	}

	// Close the file
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %v", err)
	}

	return tmpFile.Name(), nil
}
