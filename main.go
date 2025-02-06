package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"sync"

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
	cfg := parseFlags()

	// Create tempFiles to track our temporary QR code images
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
			log.Fatalf("Error generating page %d: %v", page+1, err)
		}
	}

	// Save the PDF
	if err := pdf.OutputFileAndClose(cfg.outputFile); err != nil {
		log.Fatalf("Error saving PDF: %v", err)
	}

	fmt.Printf("Generated %d pages of labels starting from %s%d\n",
		cfg.pages, cfg.prefix, cfg.startNumber)
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
