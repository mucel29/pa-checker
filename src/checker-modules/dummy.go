package checker_modules

import (
	"checker-pa/src/display"
	"fmt"
	"math/rand"
	"strconv"
)

type DummyModule struct {
	totalScore uint32
	uniqueName string
	ModuleOutput
}

func NewDummyModule() *DummyModule {
	newDummy := &DummyModule{}
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
		dummy.Issues = append(
			dummy.Issues,
			ModuleIssue{
				Message: "Lorem ipsum dolor sit amet",
				Line:    uint32(rand.Intn(255)),
				Col:     uint32(rand.Intn(100)),
			})
	}

	dummy.totalScore = uint32(rand.Intn(70))
}

func (dummy *DummyModule) Display(d *display.Display) {

	// Set the page title
	d.PrintPage(0, dummy.uniqueName, "")

	d.Println("\nTotal module score: " + strconv.Itoa(int(dummy.totalScore)))

	if len(dummy.Issues) > 0 {

		d.PrintPage(1, dummy.uniqueName+" errors", dummy.ModuleError.String())

	}
}

func (dummy *DummyModule) Dump() {
	fmt.Printf("===== %s - %d =====\n\n", dummy.uniqueName, dummy.totalScore)
	fmt.Println(dummy.ModuleError.String())
	fmt.Println()

}

func (dummy *DummyModule) Reset() {
	dummy.totalScore = 0
	dummy.Issues = nil
}

func (dummy *DummyModule) Score() uint32 {
	return dummy.totalScore
}
