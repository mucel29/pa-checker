package checkermodules

import (
	"checker-pa/src/display"
	"checker-pa/src/utils"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/sergi/go-diff/diffmatchpatch"
)

/*
const (
	MaxRow = 10
	MaxCol = 5
)
*/

// Store formatted output for side-by-side comparison
type FormattedOutput struct {
	reference []string
	output    []string
}

// Store differences for each file
type FileCompareResult struct {
	filename string
	matched  bool
	timedOut bool
	crashed  bool
	points   int
	FormattedOutput
}
type DiffModule struct {
	ModuleOutput
	totalScore int
	uniqueName string
	results    []FileCompareResult
	matchCount int
	totalFiles int
	status     ModuleStatus
}

func NewDiffModule() *DiffModule {
	return &DiffModule{
		totalScore: 0,
		uniqueName: "REFS",
	}
}

func (dm *DiffModule) GetName() string {
	return dm.uniqueName
}

func (*DiffModule) IsOutputDependent() bool {
	return utils.Config.RefChecker.OutputDependent
}

func (*DiffModule) GetDependencies() []string { return nil }

func (dm *DiffModule) Disable(fail bool) {
	if fail {
		dm.status = DependencyFail
	} else {
		dm.status = Disabled
	}
}

func (dm *DiffModule) Enable() {
	dm.status = Queued
}

func (dm *DiffModule) GetStatus() ModuleStatus {
	return dm.status
}

func (dm *DiffModule) GetResult() string {
	return fmt.Sprintf("%d / %d", dm.matchCount, dm.totalFiles)
}

func (dm *DiffModule) Panic() {
	dm.status = Panic
}

func (dm *DiffModule) Run() {
	dm.status = Running
	defer func() { dm.status = Ready }()

	config := utils.Config.UserConfig
	folder1 := utils.Abs(config.RefPath)
	folder2 := utils.Abs(config.OutputPath)

	numFiles := len(utils.Config.Tests)

	matchedCount := dm.compareFilesInFolders(folder1, folder2)
	/*
		if err != nil {
			dm.Issues = append(dm.Issues, ModuleIssue{
				Message: "Error comparing files: " + err.Error(),
			})
			return
		}
	*/

	dm.matchCount = matchedCount
	dm.totalFiles = numFiles

	// Calculate score based on matched files
	// dm.totalScore = int((float64(matchedCount) / float64(numFiles)) * 100)

	// Add issues for mismatched files
	for _, result := range dm.results {
		dm.totalScore += result.points
		switch {
		case result.timedOut:
			dm.Issues = append(dm.Issues, ModuleIssue{
				Message: fmt.Sprintf("File %s timed out", result.filename),
			})
		case result.crashed:
			dm.Issues = append(dm.Issues, ModuleIssue{
				Message: fmt.Sprintf("File %s crashed", result.filename),
			})
		case !result.matched:
			dm.Issues = append(dm.Issues, ModuleIssue{
				Message: fmt.Sprintf("File %s has differences", result.filename),
			})
		}
	}
}

func (dm *DiffModule) Display(d *display.Display) {

	// TODO: fix weird bug where selecting a cell selects 2 cells
	// TODO: maybe highlight the last clicked cell
	// TODO: this means to move the table selector back here to access the closure variables

	d.CurrentContainer().Title("Ref checker - "+strconv.Itoa(dm.totalScore), tview.AlignLeft)

	if statusStr := StatusStr(dm); statusStr != "" {
		d.PrintPage(0, "$nb", statusStr)
		return
	}

	if len(dm.Issues) == 0 {
		d.PrintPage(0, "$nb", "")
		d.CurrentContainer().Print("All files matched!")
		return
	}

	fileTable := tview.NewTable()

	fileTable.SetInputCapture(utils.TableSelector(len(dm.results), fileTable))

	currentRow := 0
	currentCol := 0

	cMaxRow, cMaxCol := utils.ComputeBestArea(len(dm.results))

	for _, result := range dm.results {
		// utils.Log(result.filename)
		if currentRow >= cMaxRow && currentCol < cMaxCol {
			currentRow = 0
			currentCol++
		}

		var cellText string
		switch {
		case result.timedOut:
			cellText = fmt.Sprintf(tview.Escape("[TO] %s"), result.filename)
		case result.crashed:
			cellText = fmt.Sprintf(tview.Escape("[SF] %s"), result.filename)
		default:
			cellText = fmt.Sprintf("[%02d] %s", result.points, result.filename)
		}

		cell := tview.NewTableCell(cellText)

		switch {
		case result.timedOut:
			cell.SetTextColor(tcell.ColorYellow)
		case result.crashed:
			cell.SetTextColor(tcell.ColorRed)
		case result.matched:
			cell.SetTextColor(tcell.ColorGreen)
		default:
			cell.SetTextColor(tcell.ColorRed)
		}

		cell.SetSelectable(true)
		cell.SetClickedFunc(func() bool {

			d.NewPage("", true)
			d.CurrentContainer().SetDirection(tview.FlexColumn)
			d.CurrentContainer().SyncSections(true)
			d.AddWritableContainer(d.CurrentContainer(), 0, 1)

			updateComparisonDisplay(d, result)

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

func (dm *DiffModule) Dump() {
	fmt.Printf("===== %s - %d =====\n\n", "Ref checker", dm.totalScore)

	if dm.status != Ready {
		fmt.Println("This module is disabled.")
		return
	}

	if len(dm.Issues) > 0 {
		fmt.Println(dm.ModuleError.String())
	} else {
		fmt.Println("All tests passed!")
	}
	fmt.Println()

}

// Helper function to update the comparison display with new file content
func updateComparisonDisplay(d *display.Display, result FileCompareResult) {

	d.App.SetFocus(d.CurrentContainer().Container)

	// Show file being viewed in both sections
	d.PrintPage(0, fmt.Sprintf("Reference - %s", result.filename), "")
	d.PrintPage(1, fmt.Sprintf("Output - %s", result.filename), "")

	// Prepare the reference section content
	var refContent strings.Builder
	refContent.WriteString("[::b]Reference content for " + result.filename + ":[white]\n")
	refContent.WriteString("------------------------------\n\n")

	for _, line := range result.reference {
		if line != "" {
			refContent.WriteString(line + "\n")
		} else {
			refContent.WriteString("\n")
		}
	}

	// Prepare the output section content
	var outContent strings.Builder
	outContent.WriteString("[::b]Output content for " + result.filename + ":[white]\n")
	if result.timedOut {
		outContent.WriteString("[yellow]WARNING: This test timed out! Output might be incomplete.[white]\n")
	} else if result.crashed {
		outContent.WriteString("[red]WARNING: This test crashed / segfaulted! Output might be incomplete.[white]\n")
	}
	outContent.WriteString("------------------------------\n\n")

	for _, line := range result.output {
		if line != "" {
			outContent.WriteString(line + "\n")
		} else {
			outContent.WriteString("\n")
		}
	}

	// Update each section with its content
	d.PrintPage(0, fmt.Sprintf("Reference - %s", result.filename), refContent.String())
	d.PrintPage(1, fmt.Sprintf("Output - %s", result.filename), outContent.String())

	// Wrap input over section 0 to support key scrolling
	d.CurrentContainer().WrapInput(d.CurrentContainer().Sections[0])
}

// Helper function to show side-by-side comparison (not needed anymore but kept for non-interactive mode)
/*
func showSideBySideComparison(display display.Display, result FileCompareResult) {
	if !display.IsInteractive() {
		// For non-interactive display, use the existing implementation
		var referenceText, outputText string

		referenceText = "Reference:\n"
		referenceText += "-----------------\n"

		outputText = "Output:\n"
		outputText += "-----------------\n"

		refLines := result.formattedOutput.reference
		outLines := result.formattedOutput.output
		maxLen := len(refLines)
		if len(outLines) > maxLen {
			maxLen = len(outLines)
		}

		for i := 0; i < maxLen; i++ {
			if i < len(refLines) && refLines[i] != "" {
				referenceText += refLines[i] + "\n"
			} else {
				referenceText += "\n"
			}

			if i < len(outLines) && outLines[i] != "" {
				outputText += outLines[i] + "\n"
			} else {
				outputText += "\n"
			}
		}

		// Display the sections one after another for basic display
		display.Println(referenceText)
		display.Println(outputText)
	}
	// For interactive mode, we now use updateComparisonDisplay instead
}
*/

func (dm *DiffModule) Reset() {
	if dm.status == Disabled || dm.status == DependencyFail {
		return
	}
	dm.totalScore = 0
	dm.Issues = nil
	dm.results = nil
	dm.matchCount = 0
	dm.totalFiles = 0
	dm.status = Queued
}

func (dm *DiffModule) Score() int {
	return int(float32(dm.totalScore) * utils.Config.RefChecker.Grade)
}

func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed reading file: %w", err)
	}
	return string(data), nil
}

// replaceWhitespaces substitutes invisible chars with visible placeholders for diff display.
func replaceWhitespaces(s string) string {
	s = strings.ReplaceAll(s, " ", "·")
	s = strings.ReplaceAll(s, "\t", "→   ")
	s = strings.ReplaceAll(s, "\r", "␍")
	return s
}

// formatLine strips trailing newline and wraps content in a tview color tag.
func formatLine(line string, color string) string {
	hasNewline := strings.HasSuffix(line, "\n")
	if hasNewline {
		line = strings.TrimSuffix(line, "\n")
	}

	escaped := tview.Escape(line)
	if color != "" {
		return fmt.Sprintf("[%s]%s[white]", color, escaped)
	}
	return escaped
}

// lineExists checks if a line appears anywhere in the text (used for swapped-line detection).
func lineExists(text string, target string) bool {
	targetTrimmed := strings.TrimSuffix(target, "\n")
	if targetTrimmed == "" {
		return false
	}
	lines := strings.Split(text, "\n")
	for _, l := range lines {
		if l == targetTrimmed {
			return true
		}
	}
	return false
}

const newlineSymbol = "↵"

// prepareLineStr strips trailing newline, replaces whitespace chars, and appends ↵ if needed.
func prepareLineStr(line string) string {
	hasNl := strings.HasSuffix(line, "\n")
	if hasNl {
		line = strings.TrimSuffix(line, "\n")
	}
	s := replaceWhitespaces(line)
	if hasNl {
		s += newlineSymbol
	}
	return s
}

// renderInlineDiff produces colored ref/out strings for two mismatched lines.
func renderInlineDiff(dmp *diffmatchpatch.DiffMatchPatch, del, ins string) (string, string) {
	delStr := prepareLineStr(del)
	insStr := prepareLineStr(ins)

	inlineDiffs := dmp.DiffMain(delStr, insStr, false)

	var rText, oText string
	for _, inline := range inlineDiffs {
		escaped := tview.Escape(inline.Text)
		switch inline.Type {
		case diffmatchpatch.DiffInsert:
			oText += fmt.Sprintf("[red]%s[white]", escaped)
		case diffmatchpatch.DiffDelete:
			rText += fmt.Sprintf("[red]%s[white]", escaped)
		case diffmatchpatch.DiffEqual:
			rText += fmt.Sprintf("[green]%s[white]", escaped)
			oText += fmt.Sprintf("[green]%s[white]", escaped)
		}
	}
	return rText, oText
}

// renderSoloLine formats a single unpaired line as red (missing/extra) with whitespace markers.
func renderSoloLine(line string) string {
	return fmt.Sprintf("[red]%s[white]", tview.Escape(prepareLineStr(line)))
}

// flushPending aligns pending deletes/inserts side-by-side and applies
// yellow (swapped), red (missing/extra), or inline char-level diff coloring.
func flushPending(
	dmp *diffmatchpatch.DiffMatchPatch,
	text1, text2 string,
	pendingDeletes, pendingInserts []string,
	refOut, outOut *[]string,
) {
	if len(pendingDeletes) == 0 && len(pendingInserts) == 0 {
		return
	}

	maxL := len(pendingDeletes)
	if len(pendingInserts) > maxL {
		maxL = len(pendingInserts)
	}

	for i := 0; i < maxL; i++ {
		var del, ins string
		hasDel := i < len(pendingDeletes)
		hasIns := i < len(pendingInserts)

		if hasDel {
			del = pendingDeletes[i]
		}
		if hasIns {
			ins = pendingInserts[i]
		}

		switch {
		case hasDel && hasIns:
			// Yellow if lines swapped positions; inline char diff otherwise.
			delTrimmed := strings.TrimSuffix(del, "\n")
			insTrimmed := strings.TrimSuffix(ins, "\n")
			if delTrimmed != insTrimmed &&
				lineExists(text2, del) && lineExists(text1, ins) {
				*refOut = append(*refOut, formatLine(del, "yellow"))
				*outOut = append(*outOut, formatLine(ins, "yellow"))
				continue
			}
			rText, oText := renderInlineDiff(dmp, del, ins)
			*refOut = append(*refOut, rText)
			*outOut = append(*outOut, oText)
		case hasDel:
			// Unpaired ref line: yellow if displaced, red if truly missing from output.
			if lineExists(text2, del) {
				*refOut = append(*refOut, formatLine(del, "yellow"))
			} else {
				*refOut = append(*refOut, renderSoloLine(del))
			}
			*outOut = append(*outOut, "")
		case hasIns:
			// Unpaired output line: yellow if displaced, red if extra/unexpected.
			*refOut = append(*refOut, "")
			if lineExists(text1, ins) {
				*outOut = append(*outOut, formatLine(ins, "yellow"))
			} else {
				*outOut = append(*outOut, renderSoloLine(ins))
			}
		}
	}
}

// compareWholeFile runs a block-aligned line diff and produces formatted side-by-side output.
func compareWholeFile(text1, text2 string) (bool, FormattedOutput) {
	dmp := diffmatchpatch.New()

	// Encode lines as single chars, diff those, then decode back to full lines.
	text1Lines, text2Lines, lineArray := dmp.DiffLinesToChars(text1, text2)
	diffs := dmp.DiffMain(text1Lines, text2Lines, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	matched := true
	for _, d := range diffs {
		if d.Type != diffmatchpatch.DiffEqual {
			matched = false
			break
		}
	}

	var refOut []string
	var outOut []string

	var pendingDeletes []string
	var pendingInserts []string

	// Split preserving trailing \n on each line; drop the empty trailing element.
	splitLines := func(s string) []string {
		if s == "" {
			return nil
		}
		lines := strings.SplitAfter(s, "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		return lines
	}

	equalColor := ""

	// Walk diff ops: queue deletes/inserts, flush on equal blocks for side-by-side alignment.
	for _, d := range diffs {
		lines := splitLines(d.Text)
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			flushPending(dmp, text1, text2, pendingDeletes, pendingInserts, &refOut, &outOut)
			pendingDeletes = nil
			pendingInserts = nil
			for _, line := range lines {
				formatted := formatLine(line, equalColor)
				refOut = append(refOut, formatted)
				outOut = append(outOut, formatted)
			}
		case diffmatchpatch.DiffDelete:
			pendingDeletes = append(pendingDeletes, lines...)
		case diffmatchpatch.DiffInsert:
			pendingInserts = append(pendingInserts, lines...)
		}
	}
	// Flush any remaining pending lines at end of file.
	flushPending(dmp, text1, text2, pendingDeletes, pendingInserts, &refOut, &outOut)

	return matched, FormattedOutput{
		reference: refOut,
		output:    outOut,
	}
}

type asyncResults struct {
	mu      sync.Mutex
	matches int
	Results []FileCompareResult
}

func (ar *asyncResults) add(index int, fcr FileCompareResult) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	ar.Results[index] = fcr
}

func (ar *asyncResults) inc() {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	ar.matches++
}

// TODO: resolve absolute path for the folders

func (dm *DiffModule) compareFilesInFolders(folder1, folder2 string) int {
	wg := sync.WaitGroup{}

	ar := asyncResults{}
	ar.Results = make([]FileCompareResult, len(utils.Config.Tests))

	for i, test := range utils.Config.Tests {
		// utils.Log(test.DisplayName)
		wg.Add(1)
		go func() {
			defer wg.Done()

			file1 := filepath.Join(folder1, fmt.Sprintf("%s.ref", test.File))
			file2 := filepath.Join(folder2, fmt.Sprintf("%s.out", test.File))

			// utils.Log(file1)
			// utils.Log(file2)

			// TODO: change the return into some kind of error

			text1, err := readFile(file1)
			if err != nil {
				utils.Err(fmt.Sprintf("failed reading file: %s", file1))
				return // 0, fmt.Errorf("error reading file1: %w", err)
			}

			text2, err := readFile(file2)
			if err != nil {
				utils.Err(fmt.Sprintf("failed reading file: %s", file2))
				return // 0, fmt.Errorf("error reading file2: %w", err)
			}

			matched, formattedOut := compareWholeFile(text1, text2)

			points := 0
			if matched {
				ar.inc()
				points = test.Score
			}

			utils.Log("Checked " + test.DisplayName)

			ar.add(i, FileCompareResult{
				filename:        test.DisplayName,
				matched:         matched,
				timedOut:        utils.Config.Tests[i].TimedOut,
				crashed:         utils.Config.Tests[i].Crashed,
				points:          points,
				FormattedOutput: formattedOut,
			})

		}()

	}

	wg.Wait()

	dm.results = append(dm.results, ar.Results...)

	return ar.matches
}


