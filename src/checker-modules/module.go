package checker_modules

import "strconv"

type ModuleIssue struct {
	Line    uint32
	Col     uint32
	Message string
}

type ModuleError struct {
	Details string
	Issues  []ModuleIssue
}

type ModuleOutput struct {
	Score uint32
	Error *ModuleError
}

type CheckerModule interface {
	GetName() string
	WaitingFor() []string
	Run()
	Details() ModuleOutput
	Reset()
}

func (err *ModuleError) String() string {
	message := err.Details + "\n"

	for _, issue := range err.Issues {
		message += strconv.Itoa(int(issue.Line)) + ":" + strconv.Itoa(int(issue.Col)) + " " + issue.Message + "\n"
	}

	return message
}
