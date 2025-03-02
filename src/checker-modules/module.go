package checker_modules

import (
	"checker-pa/src/display"
	"strconv"
)

type ModuleIssue struct {
	File        string
	Line        uint32
	Col         uint32
	Message     string
	ShowLineCol bool
}

type ModuleError struct {
	Details string
	Issues  []ModuleIssue
}

type ModuleOutput struct {
	Score   int32
	Error   *ModuleError
	Message []ModuleIssue
}

type CheckerModule interface {
	GetName() string
	WaitingFor() []string
	Run()
	// Details Print to the given output adapter
	Details(display display.Display)
	Reset()
	Score() uint32
}

func (err *ModuleError) String() string {
	message := err.Details + "\n"

	for _, issue := range err.Issues {
		message += "\n"
		if issue.ShowLineCol {
			message += strconv.Itoa(int(issue.Line)) + ":" + strconv.Itoa(int(issue.Col)) + " "
		}
		message += issue.Message + "\n"
	}

	return message
}
