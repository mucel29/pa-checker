package checker_modules

var AvailableModules = map[string]CheckerModule{
	"ref_checker":    &DummyModule{},
	"memory_checker": &DummyModule{},
	"style_checker":  &StyleChecker{},
	"commit_checker": &CommitChecker{},
}
