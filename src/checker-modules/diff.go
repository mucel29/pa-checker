package checkermodules

import (
	"checker-pa/src/display"
	"checker-pa/src/utils"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	MaxRow = 10
	MaxCol = 5
)

// Store formatted output for side-by-side comparison
type FormattedOutput struct {
	reference []string
	output    []string
}

// Store differences for each file
type FileCompareResult struct {
	filename string
	matched  bool
	diffs    []diffmatchpatch.Diff
	FormattedOutput
}
type DiffModule struct {
	ModuleOutput
	totalScore int
	uniqueName string
	results    []FileCompareResult
	matchCount int
	totalFiles int
}

func NewDiffModule() *DiffModule {
	return &DiffModule{
		totalScore: 0,
		uniqueName: "diff_checker",
	}
}

func (dm *DiffModule) GetName() string {
	return dm.uniqueName
}

func (dm *DiffModule) WaitingFor() []string {
	return []string{}
}

func (dm *DiffModule) Run() {
	config := utils.Config.UserConfig
	folder1 := config.RefPath
	folder2 := config.OutputPath

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
	dm.totalScore = int((float64(matchedCount) / float64(numFiles)) * 100)

	// Add issues for mismatched files
	for _, result := range dm.results {
		if !result.matched {
			dm.Issues = append(dm.Issues, ModuleIssue{
				Message: fmt.Sprintf("File %s has differences", result.filename),
			})
		}
	}
}

func (dm *DiffModule) Display(d *display.Display) {

	d.CurrentContainer().Title("Ref checker - "+strconv.Itoa(int(dm.totalScore)), tview.AlignLeft)

	if len(dm.Issues) > 0 {

		fileTable := tview.NewTable()

		fileTable.SetInputCapture(utils.TableSelector(len(dm.results), fileTable))

		currentRow := 0
		currentCol := 0

		for _, result := range dm.results {
			if currentRow >= MaxRow && currentCol < MaxCol {
				currentRow = 0
				currentCol++
			}

			cell := tview.NewTableCell(result.filename)

			if result.matched {
				cell.SetTextColor(tcell.ColorGreen)
			} else {
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

		fileTable.GetCell(0, 0).SetBackgroundColor(tcell.ColorWhite)

		d.CurrentContainer().AddPrimitive(fileTable, true, 0, 1)

	} else {
		d.CurrentContainer().Print("All files matched!")
	}

}

func (dm *DiffModule) Dump() {
	fmt.Printf("===== %s - %d =====\n\n", "Ref checker", dm.totalScore)
	if len(dm.Issues) > 0 {
		fmt.Println(dm.ModuleError.String())
	} else {
		fmt.Println("All tests passed!")
	}
	fmt.Println()

}

// Helper function to update the comparison display with new file content
func updateComparisonDisplay(display *display.Display, result FileCompareResult) {
	// Show file being viewed in both sections
	display.PrintPage(0, fmt.Sprintf("Reference - %s", result.filename), "")
	display.PrintPage(1, fmt.Sprintf("Output - %s", result.filename), "")

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
	outContent.WriteString("------------------------------\n\n")

	for _, line := range result.output {
		if line != "" {
			outContent.WriteString(line + "\n")
		} else {
			outContent.WriteString("\n")
		}
	}

	// Update each section with its content
	display.PrintPage(0, fmt.Sprintf("Reference - %s", result.filename), refContent.String())
	display.PrintPage(1, fmt.Sprintf("Output - %s", result.filename), outContent.String())
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
	dm.totalScore = 0
	dm.Issues = nil
	dm.results = nil
	dm.matchCount = 0
	dm.totalFiles = 0
}

func (dm *DiffModule) Score() int {
	return dm.totalScore
}

func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed reading file: %w", err)
	}
	return string(data), nil
}

func generateFormattedOutput(diffs []diffmatchpatch.Diff) FormattedOutput {
	var refText, outText string
	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			outText += fmt.Sprintf("\033[32m%s\033[0m", diff.Text)
			refText += strings.Repeat(" ", len(diff.Text))
		case diffmatchpatch.DiffDelete:
			refText += fmt.Sprintf("\033[31m%s\033[0m", diff.Text)
			outText += strings.Repeat(" ", len(diff.Text))
		case diffmatchpatch.DiffEqual:
			refText += diff.Text
			outText += diff.Text
		}
	}

	return FormattedOutput{
		reference: strings.Split(refText, "\n"),
		output:    strings.Split(outText, "\n"),
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

func (dm *DiffModule) compareFilesInFolders(folder1, folder2 string) int {
	wg := sync.WaitGroup{}

	ar := asyncResults{}
	ar.Results = make([]FileCompareResult, len(utils.Config.Tests))

	for i, test := range utils.Config.Tests {
		wg.Add(1)
		go func() {
			defer wg.Done()

			file1 := fmt.Sprintf("%s/%s.ref", folder1, test.DisplayName)
			file2 := fmt.Sprintf("%s/%s.out", folder2, test.DisplayName)

			// utils.Log(file1)

			// TODO: change the return into some kind of error

			text1, err := readFile(file1)
			if err != nil {
				return // 0, fmt.Errorf("error reading file1: %w", err)
			}

			text2, err := readFile(file2)
			if err != nil {
				return // 0, fmt.Errorf("error reading file2: %w", err)
			}

			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(text1, text2, false)

			matched := len(diffs) == 1 && diffs[0].Type == diffmatchpatch.DiffEqual
			if matched {
				ar.inc()
			}

			ar.add(i, FileCompareResult{
				filename:        test.DisplayName,
				matched:         matched,
				diffs:           diffs,
				FormattedOutput: generateFormattedOutput(diffs),
			})

		}()

	}

	wg.Wait()

	dm.results = append(dm.results, ar.Results...)

	return ar.matches
}

/*
func showDifferences(result FileCompareResult, displayType int) {
	if result.matched {
		fmt.Printf("\033[32mFile %s: Files are identical\033[0m\n", result.filename)
		return
	}

	fmt.Printf("\033[31mFile %s: Files are different\033[0m\n", result.filename)

	if displayType == 1 {
		// Original inline display
		for _, diff := range result.diffs {
			switch diff.Type {
			case diffmatchpatch.DiffInsert:
				fmt.Printf("\033[32m%s\033[0m", diff.Text)
			case diffmatchpatch.DiffDelete:
				fmt.Printf("\033[31m%s\033[0m", diff.Text)
			case diffmatchpatch.DiffEqual:
				fmt.Printf("%s", diff.Text)
			}
		}
	} else if displayType == 2 {
		fmt.Println("\nReference:")
		fmt.Println("----------")
		for _, line := range result.reference {
			if line != "" {
				fmt.Println(line)
			}
		}

		fmt.Println("\nOutput:")
		fmt.Println("-------")
		for _, line := range result.output {
			if line != "" {
				fmt.Println(line)
			}
		}
	}
	fmt.Println()
}
*/
