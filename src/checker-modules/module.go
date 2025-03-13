package checkermodules

import (
	"checker-pa/src/display"
	"strconv"
)

type ModuleIssue struct {
	File        string
	Line        int
	Col         int
	Message     string
	ShowLineCol bool
	Critical    bool
}

type ModuleError struct {
	ErrorMessage string
	Issues       []ModuleIssue
}

type ModuleOutput struct {
	Score int32
	ModuleError
}

type CheckerModule interface {
	GetName() string
	IsOutputDependent() bool
	Run()
	Display(d *display.Display)
	Dump()
	Reset()
	Score() int
}

func (err *ModuleError) String() string {
	message := err.ErrorMessage + "\n"

	for _, issue := range err.Issues {
		message += "\n"
		if issue.ShowLineCol {
			message += strconv.Itoa(int(issue.Line)) + ":" + strconv.Itoa(int(issue.Col)) + " "
		}
		message += issue.Message + "\n"
	}

	return message
}

func (err *ModuleError) groupIssues(groupBy func(issue *ModuleIssue) string) map[string][]ModuleIssue {

	group := make(map[string][]ModuleIssue)

	for _, issue := range err.Issues {
		group[groupBy(&issue)] = append(group[groupBy(&issue)], issue)
	}

	return group
}
