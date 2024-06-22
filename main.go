package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type ErrorMetric struct {
	Count      int
	Timestamps []time.Time
}

func main() {
	filePath := "server.log"
	outputFileName := "error_metrics_report.pdf"

	// Step 1: Parse the log file and extract error occurrences
	metrics, err := parseLogFile(filePath)
	if err != nil {
		fmt.Println("Error parsing log file:", err)
		return
	}

	// Step 2: Generate graphs for each error type
	for errorType, metric := range metrics {
		err := generateGraph(errorType, metric.Timestamps)
		if err != nil {
			fmt.Printf("Error generating graph for %s: %v\n", errorType, err)
			continue
		}
	}

	// Step 3: Generate the PDF report
	err = generatePDF(metrics, outputFileName)
	if err != nil {
		fmt.Println("Error generating PDF report:", err)
		return
	}

	fmt.Printf("Error metrics report has been generated and saved to %s\n", outputFileName)
}

// parseLogFile reads the log file and extracts error occurrences
func parseLogFile(filePath string) (map[string]*ErrorMetric, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	errorPattern := regexp.MustCompile(`(?i)(ERROR|EXCEPTION) .*`)
	timestampPattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)

	scanner := bufio.NewScanner(file)
	metrics := make(map[string]*ErrorMetric)

	for scanner.Scan() {
		line := scanner.Text()
		if errorPattern.MatchString(line) {
			errorType := strings.ToUpper(errorPattern.FindString(line))
			timestampStr := timestampPattern.FindString(line)
			timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
			if err != nil {
				fmt.Printf("Error parsing timestamp in line: %s\n", line)
				continue
			}

			if _, ok := metrics[errorType]; !ok {
				metrics[errorType] = &ErrorMetric{}
			}
			metrics[errorType].Count++
			metrics[errorType].Timestamps = append(metrics[errorType].Timestamps, timestamp)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return metrics, nil
}

// generateGraph generates a graph for error occurrences over time

func generateGraph(errorType string, timestamps []time.Time) error {
	p := plot.New()

	p.Title.Text = fmt.Sprintf("Occurrences of %s", errorType)
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Occurrences"

	points := make(plotter.XYs, len(timestamps))
	for i, t := range timestamps {
		points[i].X = float64(t.Unix())
		points[i].Y = float64(i + 1)
	}

	// Create scatter plot for points
	s, err := plotter.NewScatter(points)
	if err != nil {
		return fmt.Errorf("error creating scatter plot: %v", err)
	}

	s.GlyphStyle.Shape = draw.CircleGlyph{}
	s.GlyphStyle.Color = plotutil.Color(1) // Set point color

	p.Add(s)
	p.Add(plotter.NewGrid())

	// Save the plot to a PNG file
	filename := fmt.Sprintf("%s_graph.png", strings.ToLower(errorType))
	if err := p.Save(6*vg.Inch, 4*vg.Inch, filename); err != nil {
		return fmt.Errorf("error saving plot: %v", err)
	}

	return nil
}

// generatePDF generates a PDF report with error metrics and graphs
func generatePDF(metrics map[string]*ErrorMetric, outputFileName string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	pdf.Cell(40, 10, "Error Metrics Report")
	pdf.Ln(20)

	for errorType, metric := range metrics {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, fmt.Sprintf("Error Type: %s", errorType))
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, fmt.Sprintf("Total Occurrences: %d", metric.Count))
		pdf.Ln(10)

		pdf.Cell(40, 10, "Occurrences:")
		pdf.Ln(10)
		for _, timestamp := range metric.Timestamps {
			pdf.Cell(40, 10, fmt.Sprintf("  - %s", timestamp.Format("2006-01-02 15:04:05")))
			pdf.Ln(10)
		}
		pdf.Ln(10)

		// Add the corresponding graph
		graphFilename := fmt.Sprintf("%s_graph.png", strings.ToLower(errorType))
		pdf.Image(graphFilename, 10, pdf.GetY(), 180, 100, false, "", 0, "")
		pdf.Ln(100)
	}

	err := pdf.OutputFileAndClose(outputFileName)
	if err != nil {
		return fmt.Errorf("error saving PDF: %v", err)
	}

	return nil
}
