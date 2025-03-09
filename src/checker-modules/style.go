package checkermodules

import (
	"bufio"
	"bytes"
	"checker-pa/src/display"
	"checker-pa/src/utils"
	"encoding/xml"
	"fmt"
	"github.com/rivo/tview"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type StyleChecker struct {
	ModuleOutput
	totalScore int
}

func (sc *StyleChecker) GetName() string {
	return "style_checker"
}

func (sc *StyleChecker) WaitingFor() []string {
	return []string{} // No dependencies
}

func (sc *StyleChecker) Display(d *display.Display) {
	// Display module summary
	d.CurrentContainer().Title("Style checker - "+strconv.Itoa(int(sc.totalScore)), tview.AlignLeft)

	// Display errors
	if len(sc.Issues) > 0 {
		err := ModuleError{
			ErrorMessage: "Style issues found in the code",
			Issues:       sc.Issues,
		}
		d.PrintPage(0, "$nb", err.String())
	}
}

func (sc *StyleChecker) Dump() {
	fmt.Printf("===== Style Checker - %d =====\n\n", sc.totalScore)
	fmt.Println(sc.ModuleError.String())
	fmt.Println()
}

func (sc *StyleChecker) Reset() {
	sc.Issues = nil
	sc.totalScore = 0
}

func (sc *StyleChecker) Score() int {
	if sc.totalScore < 0 {
		return 0
	}

	return sc.totalScore
}

func (sc *StyleChecker) Run() {
	// Check if cppcheck is installed
	if _, err := exec.LookPath("cppcheck"); err != nil {
		sc.Issues = append(sc.Issues, ModuleIssue{
			Message:     "cppcheck is not installed",
			Line:        0,
			Col:         0,
			ShowLineCol: false,
		})
		sc.totalScore = -1 // Module failure
		return
	}

	config := utils.Config.UserConfig

	args := []string{
		"--enable=all",
		"--check-level=exhaustive",
		"--xml",
		"--xml-version=2",
		"--inconclusive",
		"--suppress=missingIncludeSystem",
		"--language=c",
		config.SourcePath,
	}

	cmd := exec.Command("cppcheck", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		sc.Issues = append(sc.Issues, ModuleIssue{
			Message: fmt.Sprintf("cppcheck execution failed: %v\n%s", err, stdout.String()),
			// stdout contains the error message
			Line:        0,
			Col:         0,
			ShowLineCol: false,
		})
		sc.totalScore = -1 // Module failure
		return
	}

	var results utils.CppcheckResults
	if err := xml.Unmarshal(stderr.Bytes(), &results); err != nil {
		sc.Issues = append(sc.Issues, ModuleIssue{
			File:        "",
			Message:     fmt.Sprintf("Failed to parse cppcheck output: %v", err),
			Line:        0,
			Col:         0,
			ShowLineCol: false,
		})
		sc.totalScore = 0
		return
	}

	// Convert cppcheck errors to module issues
	for _, err := range results.Errors {
		for _, loc := range err.Locations {
			// Read the line content from the file and create pointer
			severityColor := sc.getSeverityColor(err.Severity)
			lineWithPointer, readErr := sc.readLineAndCreatePointer(loc.File, loc.Line, loc.Column, severityColor)

			var message string
			if readErr != nil {
				// Should never be reached
				message = fmt.Sprintf("[%s] %s at %s:%d:%d", err.Severity, err.Msg, loc.File, loc.Line, loc.Column)
			} else {
				severityColor := sc.getSeverityColor(err.Severity)
				idColor := color.New(color.FgHiBlack)
				message = fmt.Sprintf("%s:%d:%d: %s: %s %s\n%s",
					loc.File,
					loc.Line,
					loc.Column,
					severityColor.Add(color.Bold).Sprint(err.Severity),
					err.Msg,
					idColor.Sprintf("[%s]", err.ID),
					lineWithPointer)
			}

			sc.Issues = append(sc.Issues, ModuleIssue{
				File:        loc.File,
				Line:        int(loc.Line),
				Col:         int(loc.Column),
				Message:     message,
				ShowLineCol: false,
			})
		}
	}

	// Calculate the score
	sc.calculateScore()
}

// TODO: Figure out a better way to calculate the score,
// maybe based on severity and type of issues found
// also these should be configurable in the config file
func (sc *StyleChecker) calculateScore() {
	baseScore := 100
	deduction := len(sc.Issues) * 5 // Deduct 5 points per issue
	sc.totalScore = int(math.Max(0, float64(baseScore-deduction)))
}

// Implementation of error formatting inspired by
// https://github.com/danmar/cppcheck/blob/main/lib/errorlogger.cpp
func (sc *StyleChecker) readLineAndCreatePointer(filePath string, lineNum int, column int, severityColor *color.Color) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		if currentLine == lineNum {
			line := scanner.Text()

			// Trim trailing whitespace
			line = strings.TrimRightFunc(line, func(r rune) bool {
				return r == '\r' || r == '\n' || r == '\t' || r == ' '
			})

			// Replace tabs with spaces
			line = strings.ReplaceAll(line, "\t", " ")

			// Ensure column is at least 1 to avoid negative Repeat count
			safeColumn := column
			if safeColumn < 1 {
				safeColumn = 1
			}

			// Create the pointer line
			pointerLine := strings.Repeat(" ", safeColumn-1) + "^"

			// Return the code line followed by pointer line
			return line + "\n" + severityColor.Sprint(pointerLine), nil
		}
	}

	return "", fmt.Errorf("line %d not found", lineNum)
}

func (sc *StyleChecker) getSeverityColor(severity string) *color.Color {
	switch severity {
	case "error":
		return color.New(color.FgRed)
	case "warning":
		return color.New(color.FgYellow)
	case "style":
		return color.New(color.FgCyan)
	case "performance":
		return color.New(color.FgMagenta)
	case "portability":
		return color.New(color.FgMagenta)
	default:
		return color.New(color.Reset)
	}
}
