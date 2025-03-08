package checker_modules

import (
	"checker-pa/src/display"
	"checker-pa/src/utils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CommitChecker struct {
	ModuleOutput
	commits []string
	score   uint32
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
	cc.Issues = nil
	cc.score = 0
}

func (cc *CommitChecker) Score() uint32 {
	return cc.score
}

func (cc *CommitChecker) Display(d *display.Display) {
	points := int32(utils.Config.CommitChecker.Score)

	d.PrintPage(0, "Commit checker summary\n", "")

	if len(cc.Issues) == 0 {
		d.Println("No issues found!")
		d.Println(fmt.Sprintf("Got %d/%d points! Congrats :)", points, points))
		return
	}

	//means we have a internal error/
	if len(cc.Issues) == 1 {
		//means we have internal error or .git doesn't exit
		if cc.Issues[0].Critical {
			d.Println("Got an error!")
			d.Println(cc.Issues[0].Message)
			return
		}
	}

	d.Println("Detected some issues!")
	for _, issue := range cc.Issues {
		d.Println(issue.Message)
	}

	d.Println(fmt.Sprintf("The final score is %d/%d.", cc.score, points))
}

func (cc *CommitChecker) Dump() {
	fmt.Printf("===== Commit checker - %d =====\n\n", cc.score)
	fmt.Println(cc.ModuleError.String())
	fmt.Println()
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
	args := []string{"log", "--oneline", "--all"}
	cmd := exec.Command("git", args...)

	output, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			issue := ModuleIssue{Message: ErrNotFound.Error(), Critical: true}
			cc.Issues = append(cc.Issues, issue)
			return
		}
		//if the student didn't "git init" before, this will give an ambiguous error
		_, newErr := os.Stat(".git")
		if errors.Is(newErr, os.ErrNotExist) {
			errMsg := "Couldn't find any commits, are you sure you ran 'git init' firstly?"
			issue := ModuleIssue{Message: errMsg, Critical: true}
			cc.Issues = append(cc.Issues, issue)
			return
		}

		issue := ModuleIssue{Message: "CRITICAL ERROR! " + err.Error(), Critical: true}
		cc.Issues = append(cc.Issues, issue)
		return
	}

	if len(output) == 0 {
		errMsg := "The checker couldn't find any commits!"
		issue := ModuleIssue{Message: errMsg}
		cc.Issues = append(cc.Issues, issue)
		return
	}

	lines := strings.Split(string(output), "\n")

	//sanity check, it shouldn't happen ... i hope
	if len(lines) == 0 {
		//maybe put something more ... non screaming
		errMsg := "CRITICAL ERROR IN COMMIT CHECKER! Please make contact the team that made the checker. #1"
		issue := ModuleIssue{Message: errMsg, Critical: true}
		cc.Issues = append(cc.Issues, issue)
		return
	}

	lines = lines[0 : len(lines)-1]
	cc.commits = make([]string, len(lines))

	for i, line := range lines {
		splitLine := strings.SplitN(line, " ", 2)

		//this shouldn't happen but ... you never know
		if len(splitLine) == 0 {
			errMsg := "CRITICAL ERROR IN COMMIT CHECKER! Please make contact the team that made the checker. #2"
			issue := ModuleIssue{Message: errMsg, Critical: true}
			cc.Issues = append(cc.Issues, issue)
			return
		}
		if utils.Config.CommitChecker.UseFormat {
			err := checkCommits(splitLine[1])
			if err != nil {
				errMsg := "Bad commit detected: " + err.Error() + " the commit was \"" + splitLine[1] + "\"\n"

				//TODO: modify deduction
				issue := ModuleIssue{Message: errMsg}
				cc.Issues = append(cc.Issues, issue)
				continue
			}
		}

		cc.commits[i] = splitLine[1]
	}

	minCommits := utils.Config.CommitChecker.MinCommits
	points := int32(utils.Config.CommitChecker.Score)
	cc.score = uint32(points)

	if minCommits > len(cc.commits) {
		pointsToDeduct := 1
		cc.score -= uint32(pointsToDeduct)
		issueMsg := "Not enough commits have been made."
		cc.Issues = append(cc.Issues, ModuleIssue{Message: issueMsg})
	}

	deduction := 2

	if points-int32(len(cc.Issues)*deduction) <= 0 {
		cc.score = 0
	} else {
		cc.score -= uint32(len(cc.Issues) * deduction)
	}
}
