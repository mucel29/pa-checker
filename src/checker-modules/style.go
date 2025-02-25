package checker_modules

import (
	"bufio"
	"bytes"
	"checker-pa/src/utils"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

type StyleChecker struct {
	issues     []ModuleIssue
	totalScore uint32
}

type cppcheckResults struct {
	XMLName xml.Name   `xml:"results"`
	Version string     `xml:"version,attr"`
	Errors  []cppError `xml:"errors>error"`
}

type cppError struct {
	ID        string     `xml:"id,attr"`
	Severity  string     `xml:"severity,attr"`
	Msg       string     `xml:"msg,attr"`
	Verbose   string     `xml:"verbose,attr"`
	Locations []location `xml:"location"`
}

type location struct {
	File   string `xml:"file,attr"`
	Line   int    `xml:"line,attr"`
	Column int    `xml:"column,attr"`
	Info   string `xml:"info,attr"`
}

func (sc *StyleChecker) GetName() string {
	return "style_checker"
}

func (sc *StyleChecker) WaitingFor() []string {
	return []string{} // No dependencies
}

func (sc *StyleChecker) Details() ModuleOutput {
	var err *ModuleError = nil
	if len(sc.issues) > 0 {
		err = &ModuleError{
			Details: "Style issues found in the code",
			Issues:  sc.issues,
		}
	}
	return ModuleOutput{
		Score:   sc.totalScore,
		Error:   err,
		Message: sc.issues,
	}
}

func (sc *StyleChecker) Reset() {
	sc.issues = nil
	sc.totalScore = 0
}

func (sc *StyleChecker) Run() {
	// Check if cppcheck is installed
	if _, err := exec.LookPath("cppcheck"); err != nil {
		sc.issues = append(sc.issues, ModuleIssue{
			Message:     "cppcheck is not installed",
			Line:        0,
			Col:         0,
			ShowLineCol: false,
		})
		sc.totalScore = 0
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
		sc.issues = append(sc.issues, ModuleIssue{
			Message:     fmt.Sprintf("cppcheck execution failed: %v\n%s", err, stdout.String()), // in stdout there is the error message
			Line:        0,
			Col:         0,
			ShowLineCol: false,
		})
		sc.totalScore = 0
		return
	}

	var results cppcheckResults
	if err := xml.Unmarshal(stderr.Bytes(), &results); err != nil {
		sc.issues = append(sc.issues, ModuleIssue{
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
			// Read the line content from the file
			lineContent, readErr := sc.readLineFromFile(loc.File, loc.Line)
			pointer := sc.createPointer(loc.Column)

			var message string
			if readErr != nil {
				// Should never be reached
				message = fmt.Sprintf("[%s] %s at %s:%d:%d", err.Severity, err.Msg, loc.File, loc.Line, loc.Column)
			} else {
				severityColor := sc.getSeverityColor(err.Severity)
				idColor := color.New(color.FgHiBlack)
				message = fmt.Sprintf("%s:%d:%d: %s: %s %s\n%s\n%s",
					loc.File,
					loc.Line,
					loc.Column,
					severityColor.Add(color.Bold).Sprint(err.Severity),
					err.Msg,
					idColor.Sprintf("[%s]", err.ID),
					lineContent,
					severityColor.Sprint(pointer))
			}

			sc.issues = append(sc.issues, ModuleIssue{
				File:        loc.File,
				Line:        uint32(loc.Line),
				Col:         uint32(loc.Column),
				Message:     message,
				ShowLineCol: false,
			})
		}
	}

	// Calculate score based on number and severity of issues
	sc.calculateScore()
}

// TODO: Figure out a better way to calculate the score,
// maybe based on severity and type of issues found
// also these should be configurable in the config file
func (sc *StyleChecker) calculateScore() {
	baseScore := uint32(100)
	deduction := uint32(len(sc.issues) * 5) // Deduct 5 points per issue
	if deduction > baseScore {
		sc.totalScore = 0
	} else {
		sc.totalScore = baseScore - deduction
	}
}

func (sc *StyleChecker) readLineFromFile(filePath string, lineNum int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close() // Ensure the file is closed after reading

	scanner := bufio.NewScanner(file)
	currentLine := 0
	for scanner.Scan() {
		currentLine++
		if currentLine == lineNum {
			return scanner.Text(), nil
		}
	}
	return "", fmt.Errorf("line %d not found", lineNum)
}

// Helper function to create the pointer string
func (sc *StyleChecker) createPointer(column int) string {
	// should never be reached
	if column < 1 {
		column = 1
	}
	if column > 1000 {
		return "" // no pointer for very long lines
	}
	return strings.Repeat(" ", column-1) + "^"
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
