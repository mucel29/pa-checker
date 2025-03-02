package checker_modules

import (
	"checker-pa/src/display"
	"fmt"
	"math/rand"
	"strconv"
)

type DummyModule struct {
	issues     []ModuleIssue
	totalScore uint32
	uniqueName string
}

func NewDummyModule() *DummyModule {
	newDummy := &DummyModule{}
	newDummy.issues = make([]ModuleIssue, 0)
	newDummy.totalScore = 0
	newDummy.uniqueName = "dummy-" + fmt.Sprintf("%x", rand.Intn(255))

	return newDummy
}

func (dummy *DummyModule) GetName() string {
	return dummy.uniqueName
}

func (dummy *DummyModule) WaitingFor() []string {
	return []string{}
}

func (dummy *DummyModule) Run() {
	const issueCount = 25

	for i := 0; i < issueCount; i++ {
		dummy.issues = append(
			dummy.issues,
			ModuleIssue{
				Message: "Lorem ipsum dolor sit amet",
				Line:    uint32(rand.Intn(255)),
				Col:     uint32(rand.Intn(100))})
	}

	dummy.totalScore = uint32(rand.Intn(70))
}

func (dummy *DummyModule) Details(display display.Display) {

	// Set the page title
	display.PrintPage(0, dummy.uniqueName, "")

	display.Println("\nTotal module score: " + strconv.Itoa(int(dummy.totalScore)))

	if len(dummy.issues) > 0 {
		err := &ModuleError{
			Details: "Lorem ipsum dolor sit amet dolor sit amet dolor sit amet",
			Issues:  dummy.issues,
		}

		display.PrintPage(1, dummy.uniqueName+" errors", err.String())

	}
}

func (dummy *DummyModule) Reset() {
	dummy.totalScore = 0
	dummy.issues = nil
}

func (dummy *DummyModule) Score() uint32 {
	return dummy.totalScore
}
