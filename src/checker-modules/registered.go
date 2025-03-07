package checker_modules

var AvailableModules = map[string]CheckerModule{
	"ref_checker":    NewDiffModule(),
	"memory_checker": NewDummyModule(),
	"style_checker":  &StyleChecker{},
	"commit_checker": NewDummyModule(),
}
