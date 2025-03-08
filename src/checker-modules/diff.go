package checker_modules

import (
	"checker-pa/src/display"
	"checker-pa/src/utils"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type DiffModule struct {
	ModuleOutput
	totalScore uint32
	uniqueName string
	results    map[string]FileCompareResult
	matchCount int
	totalFiles int
}

func NewDiffModule() *DiffModule {
	return &DiffModule{
		totalScore: 0,
		uniqueName: "diff_checker",
		results:    make(map[string]FileCompareResult),
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

	numFiles, err := countFilesInFolder(folder1)
	if err != nil {
		dm.Issues = append(dm.Issues, ModuleIssue{
			Message: "Error counting files: " + err.Error(),
		})
		return
	}

	if numFiles == 0 {
		dm.Issues = append(dm.Issues, ModuleIssue{
			Message: "No files found in reference folder",
		})
		return
	}

	matchedCount, results, err := compareFilesInFolders(folder1, folder2, numFiles)
	if err != nil {
		dm.Issues = append(dm.Issues, ModuleIssue{
			Message: "Error comparing files: " + err.Error(),
		})
		return
	}

	dm.results = results
	dm.matchCount = matchedCount
	dm.totalFiles = numFiles

	// Calculate score based on matched files
	dm.totalScore = uint32((float64(matchedCount) / float64(numFiles)) * 100)

	// Add issues for mismatched files
	for fileName, result := range results {
		if !result.matched {
			dm.Issues = append(dm.Issues, ModuleIssue{
				Message: fmt.Sprintf("File %s has differences", fileName),
			})
		}
	}
}

func (dm *DiffModule) Display(d *display.Display) {
	// First page: Main summary
	d.PrintPage(0, dm.uniqueName, "")
	d.Println(fmt.Sprintf("\nMatched files: %d/%d", dm.matchCount, dm.totalFiles))
	d.Println(fmt.Sprintf("Score: %d/100", dm.totalScore))

	if len(dm.Issues) > 0 {
		// Second page: List of files with errors (selection menu)
		d.PrintPage(1, dm.uniqueName+" errors", "")
		d.Println("\nIncorrect files (select one to view differences):")
		d.Println("-----------------------------------------------")

		incorrectFiles := make([]int, 0)

		for fileName, result := range dm.results {
			if !result.matched {
				fileNum := 0
				fmt.Sscanf(fileName, "data%d.out", &fileNum)
				incorrectFiles = append(incorrectFiles, fileNum)
			}
		}

		if len(incorrectFiles) == 0 {
			d.Println("None - all files are correct!")
			return
		}

		// Sort numbers for better readability
		sort.Ints(incorrectFiles)

		// Print the list of files
		for _, num := range incorrectFiles {
			// In interactive mode, we'll make this section more visible
			// Make the entries stand out
			d.Println(fmt.Sprintf("● [yellow]data%d.out[white] (press %d in main tab to view this file)", num, num))

		}

		// For interactive mode, create separate reference and output blocks
		// Setup the reference section
		//d.PrintPage(2, "Reference", "")

		// Setup the output section
		//d.PrintPage(3, "Output", "")

		// TODO: rewrite

		// Start with the first incorrect file for demonstration
		//if len(incorrectFiles) > 0 {
		//	fileNum := incorrectFiles[0]
		//	fileName := fmt.Sprintf("data%d.out", fileNum)
		//
		//	//if result, exists := dm.results[fileName]; exists && !result.matched {
		//	//	// Clear and update the comparison sections
		//	//	updateComparisonDisplay(d, fileName, result)
		//	//}
		//}

		// Since the keyboard navigation is not fully implemented, show instructions for now
		d.Println("\nCurrently displaying the first incorrect file.")

		// Comment this section out for now until fully implemented
		/*
		   // Set up keyboard handlers for numbers to select files
		   if interactiveDisplay, ok := d.(*d.Display); ok {
		       interactiveDisplay.SetupFileSelectionHandler(func(fileNum int) {
		           fileName := fmt.Sprintf("data%d.out", fileNum)
		           if result, exists := dm.results[fileName]; exists && !result.matched {
		               updateComparisonDisplay(d, fileName, result)
		           }
		       }, incorrectFiles)
		   }
		*/
		//} else {
		//	// Basic mode remains unchanged
		//	for {
		//		d.Println("\nWhich file would you like to check? (Enter a number or 'q' to quit): ")
		//		input := d.ReadLine()
		//
		//		// Check if user wants to quit
		//		if input == "q" || input == "Q" || input == "quit" || input == "exit" {
		//			break
		//		}
		//
		//		fileNum, err := strconv.Atoi(input)
		//		if err != nil || fileNum < 1 || fileNum > dm.totalFiles {
		//			d.Println("Invalid file number. Please try again.")
		//			continue
		//		}
		//
		//		fileName := fmt.Sprintf("data%d.out", fileNum)
		//		if result, exists := dm.results[fileName]; exists {
		//			d.Println(fmt.Sprintf("\n=== %s ===\n", fileName))
		//
		//			if result.matched {
		//				d.Println("Files are identical\n")
		//			} else {
		//				// Show inline differences (type 1)
		//				for _, diff := range result.diffs {
		//					switch diff.Type {
		//					case diffmatchpatch.DiffInsert:
		//						d.Print(fmt.Sprintf("\033[32m%s\033[0m", diff.Text))
		//					case diffmatchpatch.DiffDelete:
		//						d.Print(fmt.Sprintf("\033[31m%s\033[0m", diff.Text))
		//					case diffmatchpatch.DiffEqual:
		//						d.Print(diff.Text)
		//					}
		//				}
		//				d.Println("\n")
		//			}
		//		} else {
		//			d.Println(fmt.Sprintf("File %s not found", fileName))
		//		}
		//	}
		//}
	} else {
		d.Println("\nSUCCESS! - All files match!")
	}
}

func (dm *DiffModule) Dump() {
	fmt.Printf("===== %s - %d =====\n\n", dm.uniqueName, dm.totalScore)
	fmt.Println(dm.ModuleError.String())
	fmt.Println()
}

// Helper function to update the comparison display with new file content
func updateComparisonDisplay(display display.Display, fileName string, result FileCompareResult) {
	// Show file being viewed in both sections
	display.PrintPage(2, fmt.Sprintf("Reference - %s", fileName), "")
	display.PrintPage(3, fmt.Sprintf("Output - %s", fileName), "")

	// Prepare the reference section content
	var refContent string
	refContent = "[::b]Reference content for " + fileName + ":[white]\n"
	refContent += "------------------------------\n\n"

	for _, line := range result.formattedOutput.reference {
		if line != "" {
			refContent += line + "\n"
		} else {
			refContent += "\n"
		}
	}

	// Prepare the output section content
	var outContent string
	outContent = "[::b]Output content for " + fileName + ":[white]\n"
	outContent += "------------------------------\n\n"

	for _, line := range result.formattedOutput.output {
		if line != "" {
			outContent += line + "\n"
		} else {
			outContent += "\n"
		}
	}

	// Update each section with its content
	display.PrintPage(2, fmt.Sprintf("Reference - %s", fileName), refContent)
	display.PrintPage(3, fmt.Sprintf("Output - %s", fileName), outContent)
}

//// Helper function to show side-by-side comparison (not needed anymore but kept for non-interactive mode)
//func showSideBySideComparison(display display.Display, result FileCompareResult) {
//	if !display.IsInteractive() {
//		// For non-interactive display, use the existing implementation
//		var referenceText, outputText string
//
//		referenceText = "Reference:\n"
//		referenceText += "-----------------\n"
//
//		outputText = "Output:\n"
//		outputText += "-----------------\n"
//
//		refLines := result.formattedOutput.reference
//		outLines := result.formattedOutput.output
//		maxLen := len(refLines)
//		if len(outLines) > maxLen {
//			maxLen = len(outLines)
//		}
//
//		for i := 0; i < maxLen; i++ {
//			if i < len(refLines) && refLines[i] != "" {
//				referenceText += refLines[i] + "\n"
//			} else {
//				referenceText += "\n"
//			}
//
//			if i < len(outLines) && outLines[i] != "" {
//				outputText += outLines[i] + "\n"
//			} else {
//				outputText += "\n"
//			}
//		}
//
//		// Display the sections one after another for basic display
//		display.Println(referenceText)
//		display.Println(outputText)
//	}
//	// For interactive mode, we now use updateComparisonDisplay instead
//}

func (dm *DiffModule) Reset() {
	dm.totalScore = 0
	dm.Issues = nil
	dm.results = make(map[string]FileCompareResult)
	dm.matchCount = 0
	dm.totalFiles = 0
}

func (dm *DiffModule) Score() uint32 {
	return dm.totalScore
}

// Store formatted output for side-by-side comparison
type FormattedOutput struct {
	reference []string
	output    []string
}

// Store differences for each file
type FileCompareResult struct {
	matched         bool
	diffs           []diffmatchpatch.Diff
	formattedOutput FormattedOutput
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

func compareFilesInFolders(folder1, folder2 string, numFiles int) (int, map[string]FileCompareResult, error) {
	matchedCount := 0
	results := make(map[string]FileCompareResult)

	for i := 1; i <= numFiles; i++ {
		fileName := fmt.Sprintf("data%d.out", i)
		file1 := fmt.Sprintf("%s/%s", folder1, fileName)
		file2 := fmt.Sprintf("%s/%s", folder2, fileName)

		text1, err := readFile(file1)
		if err != nil {
			return 0, nil, fmt.Errorf("error reading file1: %w", err)
		}

		text2, err := readFile(file2)
		if err != nil {
			return 0, nil, fmt.Errorf("error reading file2: %w", err)
		}

		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(text1, text2, false)

		matched := len(diffs) == 1 && diffs[0].Type == diffmatchpatch.DiffEqual
		if matched {
			matchedCount++
		}

		results[fileName] = FileCompareResult{
			matched:         matched,
			diffs:           diffs,
			formattedOutput: generateFormattedOutput(diffs),
		}
	}

	return matchedCount, results, nil
}

func showDifferences(fileName string, result FileCompareResult, displayType int) {
	if result.matched {
		fmt.Printf("\033[32mFile %s: Files are identical\033[0m\n", fileName)
		return
	}

	fmt.Printf("\033[31mFile %s: Files are different\033[0m\n", fileName)

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
		for _, line := range result.formattedOutput.reference {
			if line != "" {
				fmt.Println(line)
			}
		}

		fmt.Println("\nOutput:")
		fmt.Println("-------")
		for _, line := range result.formattedOutput.output {
			if line != "" {
				fmt.Println(line)
			}
		}
	}
	fmt.Println()
}

func countFilesInFolder(folder string) (int, error) {
	pattern := filepath.Join(folder, "data*.out")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return 0, fmt.Errorf("error counting files: %w", err)
	}
	return len(matches), nil
}

func showIncorrectFiles(results map[string]FileCompareResult) {
	fmt.Println("\nIncorrect files:")
	fmt.Println("---------------")
	hasIncorrect := false

	// Sort the files numerically
	incorrectFiles := make([]int, 0)
	for fileName, result := range results {
		if !result.matched {
			fileNum := 0
			fmt.Sscanf(fileName, "data%d.out", &fileNum)
			incorrectFiles = append(incorrectFiles, fileNum)
			hasIncorrect = true
		}
	}

	if !hasIncorrect {
		fmt.Println("None - all files are correct!")
		return
	}

	// Sort numbers for better readability
	sort.Ints(incorrectFiles)

	// Print in a clean format
	for _, num := range incorrectFiles {
		fmt.Printf("- data%d.out\n", num)
	}

	fmt.Println()
}
