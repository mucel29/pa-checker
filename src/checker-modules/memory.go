package checker_modules

import (
	"checker-pa/src/display"
	"checker-pa/src/utils"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	definitelyLeaked = "Leak_DefinitelyLost"
)

// ValgrindOutput represents a simplified version of Valgrind XML output focused on errors
type ValgrindOutput struct {
	Errors []Error `xml:"error"`
}

// Error represents a single error detected by Valgrind
type Error struct {
	Kind    string `xml:"kind"`
	What    string `xml:"what,omitempty"`  // Regular error description
	XWhat   XWhat  `xml:"xwhat,omitempty"` // Extended error description (for leaks)
	Stack   Stack  `xml:"stack"`
	AuxWhat string `xml:"auxwhat,omitempty"` // Additional error information
}

func (err *Error) isUserGenerated() bool {
	return strings.Contains(err.Stack.Frames[0].Obj, "/home/")
}

// XWhat contains simplified extended error information for memory leaks
type XWhat struct {
	Text        string `xml:"text"`
	LeakedBytes int    `xml:"leakedbytes"`
}

// Stack represents a call stack for an error
type Stack struct {
	Frames []Frame `xml:"frame"`
}

// Frame represents a simplified single stack frame
type Frame struct {
	Fn   string `xml:"fn"`
	File string `xml:"file,omitempty"`
	Line int    `xml:"line,omitempty"`
	Obj  string `xml:"obj,omitempty"`
}

type memoryCheckerIssue struct {
	message    string
	function   string
	file       string
	line       int
	isCritical bool
}

func (mci *memoryCheckerIssue) String() string {
	str := mci.file + ":" + strconv.Itoa(mci.line) + " inside " + mci.function + " "
	str += mci.message

	return str
}

type MemoryChecker struct {
	score    uint32
	issues   []memoryCheckerIssue
	warnings []memoryCheckerIssue
}

func (*MemoryChecker) GetName() string {
	return "memory_checker"
}

func (*MemoryChecker) WaitingFor() []string {
	return utils.Config.MemoryChecker.RunAfter
}

func (mc *MemoryChecker) Reset() {
	mc.issues = nil
	mc.score = 0
}

func (mc *MemoryChecker) Score() uint32 {
	return mc.score
}

func (mc *MemoryChecker) warningsString() string {
	warnMsg := "Found some warnings!" + "\n"
	for _, warning := range mc.warnings {
		warnMsg += warning.String() + "\n"
	}

	return warnMsg
}

func (mc *MemoryChecker) issuesString() string {
	issueMsg := "Found issues!" + "\n"
	for _, issue := range mc.issues {
		issueMsg += issue.String() + "\n"
	}

	return issueMsg
}

func (mc *MemoryChecker) Display(d *display.Display) {
	d.PrintPage(0, "Memory checker summary\n", "")

	if len(mc.issues) == 1 && mc.issues[0].isCritical {
		d.Println("Critical error detected!")
		d.Println(mc.issues[0].message)
		return
	}

	if len(mc.issues) == 0 && len(mc.warnings) == 0 {
		d.Println(
			fmt.Sprintf("No issues found! Great job you got %d/%d :)!",
				mc.score, mc.score))
		return
	}

	if len(mc.issues) == 0 {

		d.Println(mc.warningsString())

		d.Println(
			fmt.Sprintf("Your score is %d/%d!",
				mc.score, utils.Config.MemoryChecker.Score))
		return
	}

	//d.Println("Found issues!")
	//for _, issue := range mc.issues {
	//	d.Println(issue.String())
	//}

	d.Println(mc.issuesString())

	if len(mc.warnings) != 0 {
		d.Println(mc.warningsString())

		//d.Println("Found some warnings!")
		//for _, warning := range mc.warnings {
		//	errMsg := warning.file + ":" + strconv.Itoa(warning.line) + " inside " + warning.function + " "
		//	errMsg += warning.message
		//	d.Println(errMsg)
		//}
	}

	d.Println(
		fmt.Sprintf("Your score is %d/%d!",
			mc.score, utils.Config.MemoryChecker.Score))
	return
}

func (mc *MemoryChecker) Dump() {
	fmt.Printf("===== %s - %d =====\n\n", "Memory checker", mc.score)

	if len(mc.issues) == 1 && mc.issues[0].isCritical {
		fmt.Println("Critical error detected!")
		fmt.Println(mc.issues[0].message)
		return
	}

	if len(mc.warnings) > 0 {
		fmt.Println(mc.warningsString())
	}

	if len(mc.issues) > 0 {
		fmt.Println(mc.issuesString())
	}

	if len(mc.issues) == 0 && len(mc.warnings) == 0 {
		fmt.Println("No issues found! Great job :)!")
	}

	fmt.Println()
}

func (mc *MemoryChecker) Run() {
	//MOCK DATA
	//TODO: remove this later
	data := []byte{}

	data, err := os.ReadFile("./temp/foobar.xml")
	if err != nil {
		panic(err)
	}

	var output ValgrindOutput
	err = xml.Unmarshal(data, &output)
	if err != nil {
		issue := memoryCheckerIssue{message: "CRITICAL ERROR! " + err.Error(), isCritical: true}
		mc.issues = append(mc.issues, issue)
		return
	}

	mc.score = uint32(utils.Config.MemoryChecker.Score)

	idx := len(output.Errors) - 1
	for idx > -1 && output.Errors[idx].Kind == definitelyLeaked {
		mci := memoryCheckerIssue{message: output.Errors[idx].XWhat.Text}
		mci.file = output.Errors[idx].Stack.Frames[1].File
		mci.function = output.Errors[idx].Stack.Frames[1].Fn
		mci.line = output.Errors[idx].Stack.Frames[1].Line

		mc.issues = append(mc.issues, mci)

		idx--
	}

	for idx > -1 {
		if output.Errors[idx].isUserGenerated() {
			w := memoryCheckerIssue{message: output.Errors[idx].What}
			w.file = output.Errors[idx].Stack.Frames[1].File
			w.function = output.Errors[idx].Stack.Frames[1].Fn
			w.line = output.Errors[idx].Stack.Frames[1].Line

			mc.warnings = append(mc.warnings, w)
		}

		idx--
	}

	deduction := 2
	if int32(mc.score)-int32(len(mc.issues)*deduction) <= 0 {
		mc.score = 0
	} else {
		mc.score -= uint32(len(mc.issues)) * uint32(deduction)
	}

}
