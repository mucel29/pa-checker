package checker_modules

import (
	"checker-pa/src/utils"
	"errors"
	"math"
	"os"
	"os/exec"
	"strings"
)

type issue struct {
	message   string
	deduction uint32
}

type CommitChecker struct {
	issues  []issue
	commits []string
}

var ErrNotFound error = errors.New("The checker couldn't find git on your system. Are you sure it's installed?")

func (*CommitChecker) GetName() string {
	return "commit_checker"
}

func (*CommitChecker) WaitingFor() []string {
	return utils.Config.CommitChecker.RunAfter
}

func (cc *CommitChecker) Reset() {
	cc.commits = nil
	cc.issues = nil
}

func (cc *CommitChecker) Details() ModuleOutput {
	minCommits := utils.Config.CommitChecker.MinCommits
	points := uint32(utils.Config.CommitChecker.Score)
	mo := ModuleOutput{Score: points}

	if minCommits > len(cc.commits) {
		pointsToDeduct := uint32(2)
		mo.Score -= pointsToDeduct
		issueMsg := "Not enough commits have been made."
		mo.Message = append(mo.Message, ModuleIssue{Message: issueMsg, ShowLineCol: false})
	}

	if len(cc.issues) == 0 {
		return mo
	}

	//means we have a internal error/
	if len(cc.issues) == 1 {
		//means we have internal error or .git doesn't exit
		if cc.issues[0].deduction == 0 {
			mo.Error = &ModuleError{Details: cc.issues[0].message}
			mo.Score = 0
			return mo
		}
	}

	moduleIssues := make([]ModuleIssue, len(cc.issues))
	for i, issue := range cc.issues {
		mo.Score -= max(0, mo.Score-issue.deduction)
		moduleIssues[i] = ModuleIssue{Message: issue.message, ShowLineCol: false}
	}

	if len(mo.Message) == 1 {
		mo.Message = append(mo.Message, moduleIssues...)
	} else {
		mo.Message = moduleIssues
	}

	return mo
}

// receives the commit line without the commit hash
func checkCommits(line string) error {
	if !strings.Contains(line, ":") {
		return errors.New("Invalid format! Hint, the format is: <type of commit>: <message>).")
	}

	typeAndMessage := strings.SplitN(line, ":", 2)
	message := strings.Trim(typeAndMessage[1], " ")

	//TODO: replace 10 later
	if len(message) < 10 {
		return errors.New("The message is too short!")
	}

	return nil
}

func (cc *CommitChecker) Run() {
	args := []string{"log", "--oneline", "-all"}
	cmd := exec.Command("git", args...)

	output, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			issue := issue{message: ErrNotFound.Error(), deduction: math.MaxInt32}
			cc.issues = append(cc.issues, issue)
			return
		}
		//if the student didn't "git init" before, this will give an ambiguous error
		_, err := os.Stat(".git")
		if errors.Is(err, os.ErrNotExist) {
			errMsg := "Couldn't find any commits, are you sure you ran 'git init' firstly?"
			issue := issue{message: errMsg, deduction: 0}
			cc.issues = append(cc.issues, issue)
			return
		}

		issue := issue{message: "CRITICAL ERROR! " + err.Error(), deduction: 0}
		cc.issues = append(cc.issues, issue)
		return
	}

	if len(output) == 0 {
		errMsg := "The checker couldn't find any commits!"
		issue := issue{message: errMsg}
		cc.issues = append(cc.issues, issue)
		return
	}

	lines := strings.Split(string(output), "\n")

	//sanity check, it shouldn't happen ... i hope
	if len(lines) == 0 {
		//maybe put something more ... non screaming
		errMsg := "CRITICAL ERROR IN COMMIT CHECKER! #1"
		issue := issue{message: errMsg, deduction: 0}
		cc.issues = append(cc.issues, issue)
		return
	}

	lines = lines[0 : len(lines)-1]
	cc.commits = make([]string, len(lines))

	for i, line := range lines {
		splitLine := strings.SplitN(line, " ", 2)

		//this shouldn't happen but ... you never know
		if len(splitLine) == 0 {
			errMsg := "CRITICAL ERROR IN COMMIT CHECKER! #2"
			issue := issue{message: errMsg, deduction: 0}
			cc.issues = append(cc.issues, issue)
			return
		}
		if utils.Config.CommitChecker.UseFormat {
			err := checkCommits(splitLine[1])
			if err != nil {
				errMsg := "Bad commit detected: " + err.Error() + " the commit was \"" + splitLine[1] + "\"\n"

				//TODO: modify deduction
				issue := issue{message: errMsg, deduction: 1}
				cc.issues = append(cc.issues, issue)
				continue
			}
		}

		cc.commits[i] = splitLine[1]
	}
}
