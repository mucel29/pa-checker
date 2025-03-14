package checkermodules

import (
	"bufio"
	"bytes"
	"checker-pa/src/display"
	"checker-pa/src/utils"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/fatih/color"
)

type StyleChecker struct {
	ModuleOutput
	totalScore int
	status     ModuleStatus
}

func (sc *StyleChecker) GetName() string {
	return "STYLE"
}

func (sc *StyleChecker) IsOutputDependent() bool {
	return utils.Config.StyleChecker.OutputDependent
}

func (sc *StyleChecker) GetDependencies() []string { return utils.Config.StyleChecker.Dependencies }

func (sc *StyleChecker) Disable(fail bool) {
	if fail {
		sc.status = DependencyFail
	} else {
		sc.status = Disabled
	}
}

func (sc *StyleChecker) Enable() {
	sc.status = Ready
}

func (sc *StyleChecker) GetStatus() ModuleStatus {
	return sc.status
}

func (sc *StyleChecker) GetResult() string {
	return fmt.Sprintf("%d issues", len(sc.Issues))
}

func (sc *StyleChecker) Panic() {
	sc.status = Panic
}

func (sc *StyleChecker) Display(d *display.Display) {

	switch sc.status {
	case Disabled:
		d.Println("This module is disabled.")
		return
	case DependencyFail:
		d.Println("One or more dependencies have failed.\nCheck if you have the following installed:")
		for _, dependency := range sc.GetDependencies() {
			d.Println(dependency)
		}
		return
	case FakeRunning:
		fallthrough
	case Running:
		d.Println("This module is currently running. Please wait")
		return
	case Panic:
		d.PrintPage(0, "$nb", "")
		d.Println("The checker went into panic. Check the config and run again")
		return
	default:
	}

	// Display module summary
	d.CurrentContainer().Title("Style checker - "+strconv.Itoa(int(sc.totalScore)), tview.AlignLeft)

	// TODO: also sort issues by line number and col after grouping

	groups := sc.ModuleError.groupIssues(func(issue *ModuleIssue) string {
		return issue.File
	})

	if groups[""] != nil {
		d.PrintPage(0, "$nb", groups[""][0].Message)
		return
	}

	fileTable := tview.NewTable()

	fileTable.SetInputCapture(utils.TableSelector(len(groups), fileTable))

	currentRow := 0
	currentCol := 0

	var keys []string

	for k := range groups {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, file := range keys {
		if currentRow >= MaxRow && currentCol < MaxCol {
			currentRow = 0
			currentCol++
		}

		cell := tview.NewTableCell(file)

		cell.SetTextColor(tcell.ColorDarkCyan)

		cell.SetSelectable(true)
		cell.SetClickedFunc(func() bool {

			d.NewPage("[darkcyan]"+file, true)
			d.CurrentContainer().SetDirection(tview.FlexColumn)
			d.CurrentContainer().SyncSections(true)
			d.AddWritableContainer(d.CurrentContainer(), 0, 1)

			d.PrintPage(0, "$nb", "")

			for _, issue := range groups[file] {
				d.Println(issue.Message)
			}

			d.App.SetFocus(d.CurrentContainer().Container)
			d.CurrentContainer().WrapInput(d.CurrentContainer().Sections[0])

			return false
		})
		fileTable.SetCell(currentRow, currentCol, cell)

		currentRow++
	}

	firstCell := fileTable.GetCell(0, 0)

	textColor, _, _ := firstCell.Style.Decompose()

	// Create reverse style
	firstCell.SetBackgroundColor(textColor)
	firstCell.SetTextColor(tcell.ColorWhite)

	d.CurrentContainer().AddPrimitive(fileTable, true, 0, 1)

}

func (sc *StyleChecker) Dump() {
	fmt.Printf("===== Style Checker - %d =====\n\n", sc.totalScore)

	if sc.status != Ready {
		fmt.Println("The commit module is disabled.")
		return
	}

	fmt.Println(sc.ModuleError.String())
	fmt.Println()
}

func (sc *StyleChecker) Reset() {
	if sc.status == Disabled || sc.status == DependencyFail {
		return
	}
	sc.Issues = nil
	sc.totalScore = 0
	sc.status = FakeRunning
}

func (sc *StyleChecker) Score() int {
	if sc.totalScore < 0 {
		return 0
	}

	return int(float32(sc.totalScore) * utils.Config.StyleChecker.Grade)
}

func (sc *StyleChecker) Run() {
	sc.status = Running
	defer func() { sc.status = Ready }()

	// Check if cppcheck is installed
	/*
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
	*/

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
