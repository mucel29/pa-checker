package checker_modules

var AvailableModules = map[string]CheckerModule{
	"ref_checker":    NewDummyModule(),
	"memory_checker": &MemoryChecker{},
	"style_checker":  &StyleChecker{},
	"commit_checker": &CommitChecker{},
}
