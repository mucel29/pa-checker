package checker_modules

import "math/rand"

type DummyModule struct {
	issues     []ModuleIssue
	totalScore int32
}

func (dummy *DummyModule) GetName() string {
	return "dummy"
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

	dummy.totalScore = int32(rand.Intn(70))
}

func (dummy *DummyModule) Details() ModuleOutput {
	var err *ModuleError = nil

	if len(dummy.issues) > 0 {
		err = &ModuleError{
			Details: "Lorem ipsum dolor sit amet dolor sit amet dolor sit amet",
			Issues:  dummy.issues,
		}
	}

	return ModuleOutput{
		Score: dummy.totalScore,
		Error: err,
	}
}

func (dummy *DummyModule) Reset() {
	dummy.totalScore = 0
	dummy.issues = nil
}
