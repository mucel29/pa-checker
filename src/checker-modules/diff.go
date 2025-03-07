package checker_modules

import (
	"checker-pa/src/display"
	"checker-pa/src/utils"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type DiffModule struct {
	issues      []ModuleIssue
	totalScore  uint32
	uniqueName  string
	results     map[string]FileCompareResult
	matchCount  int
	totalFiles  int
}

func NewDiffModule() *DiffModule {
	return &DiffModule{
		issues:     make([]ModuleIssue, 0),
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
		dm.issues = append(dm.issues, ModuleIssue{
			Message: "Error counting files: " + err.Error(),
		})
		return
	}

	if numFiles == 0 {
		dm.issues = append(dm.issues, ModuleIssue{
			Message: "No files found in reference folder",
		})
		return
	}

	matchedCount, results, err := compareFilesInFolders(folder1, folder2, numFiles)
	if err != nil {
		dm.issues = append(dm.issues, ModuleIssue{
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
			dm.issues = append(dm.issues, ModuleIssue{
				Message: fmt.Sprintf("File %s has differences", fileName),
			})
		}
	}
}

func (dm *DiffModule) Details(display display.Display) {
	display.PrintPage(0, dm.uniqueName, "")
	
	display.Println(fmt.Sprintf("\nMatched files: %d/%d", dm.matchCount, dm.totalFiles))
	display.Println(fmt.Sprintf("Score: %d/100", dm.totalScore))

	if len(dm.issues) > 0 {
		err := &ModuleError{
			Details: "Some files have differences",
			Issues:  dm.issues,
		}
		display.PrintPage(1, dm.uniqueName+" errors", err.String())
		
		// Add interactive diff viewing
		display.Println("\nWould you like to see differences for a file? (y/n): ")
		response := display.ReadLine()
		
		for response == "y" || response == "Y" {
			display.Println(fmt.Sprintf("Enter file number (1-%d): ", dm.totalFiles))
			fileNumStr := display.ReadLine()
			fileNum, err := strconv.Atoi(fileNumStr)
			
			if err != nil || fileNum < 1 || fileNum > dm.totalFiles {
				display.Println(fmt.Sprintf("Invalid file number. Please enter a number between 1 and %d", dm.totalFiles))
				continue
			}
			
			display.Println("Enter display type (1 for inline, 2 for side by side): ")
			displayTypeStr := display.ReadLine()
			displayType, err := strconv.Atoi(displayTypeStr)
			
			if err != nil || (displayType != 1 && displayType != 2) {
				display.Println("Invalid display type")
				continue
			}
			
			fileName := fmt.Sprintf("data%d.out", fileNum)
			if result, exists := dm.results[fileName]; exists {
				if result.matched {
					display.Println(fmt.Sprintf("\033[32mFile %s: Files are identical\033[0m\n", fileName))
				} else {
					display.Println(fmt.Sprintf("\033[31mFile %s: Files are different\033[0m\n", fileName))
					
					if displayType == 1 {
						for _, diff := range result.diffs {
							switch diff.Type {
							case diffmatchpatch.DiffInsert:
								display.Println(fmt.Sprintf("\033[32m%s\033[0m", diff.Text))
							case diffmatchpatch.DiffDelete:
								display.Println(fmt.Sprintf("\033[31m%s\033[0m", diff.Text))
							case diffmatchpatch.DiffEqual:
								display.Println(diff.Text)
							}
						}
					} else {
						display.Println("\nReference:")
						display.Println("----------")
						for _, line := range result.formattedOutput.reference {
							if line != "" {
								display.Println(line)
							}
						}
						
						display.Println("\nOutput:")
						display.Println("-------")
						for _, line := range result.formattedOutput.output {
							if line != "" {
								display.Println(line)
							}
						}
					}
				}
			}
			
			display.Println("\nWould you like to see another file? (y/n): ")
			response = display.ReadLine()
		}
	} else {
		display.Println("\nSUCCESS! - All files match!")
	}
}

func (dm *DiffModule) Reset() {
	dm.totalScore = 0
	dm.issues = nil
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
        if (!result.matched) {
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
