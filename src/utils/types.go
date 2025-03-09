package utils

import "encoding/xml"

type Test struct {
	DisplayName string   `json:"displayName"`
	File        string   `json:"file"`
	Args        []string `json:"args"`
	Ordered     bool     `json:"ordered"`
	WhiteSpace  bool     `json:"whitespace"`
	Score       int      `json:"score"`
}

type RefChecker struct {
	RunAfter  []string `json:"run_after"`
	InputPath string   `json:"input_after"`
	Tests     []Test   `json:"tests"`
}

type CommitChecker struct {
	RunAfter   []string `json:"run_after"`
	MinCommits int      `json:"minCommits"`
	UseFormat  bool     `json:"useFormat"`
	Score      int      `json:"score"`
}

type MemoryChecker struct {
	RunAfter   []string `json:"run_after"`
	MaxWarning int      `json:"maxWarning"`
	MaxLeak    int      `json:"maxLeak"`
	Score      int      `json:"score"`
}

type StyleChecker struct {
	RunAfter       []string `json:"run_after"`
	ScoreThreshold int      `json:"score_threshold"`
	Thresholds     []struct {
		Under int `json:"under"`
		Score int `json:"score"`
	} `json:"thresholds"`
}

type ModuleConfig struct {
	TempPath       string `json:"temp_path"`
	RunValgrind    bool   `json:"run_valgrind"`
	*RefChecker    `json:"ref_checker"`
	*CommitChecker `json:"commit_checker"`
	*MemoryChecker `json:"memory_checker"`
	*StyleChecker  `json:"style_checker"`
}

type UserConfig struct {
	SourcePath     string `json:"source_path"`
	ExecutablePath string `json:"executable_path"`
	InputPath      string `json:"input_path"`
	OutputPath     string `json:"output_path"`
	RefPath        string `json:"ref_path"`
	ForwardPath    string `json:"forward_path"`
}

type CppcheckResults struct {
	XMLName xml.Name   `xml:"results"`
	Version string     `xml:"version,attr"`
	Errors  []CppError `xml:"errors>error"`
}

type CppError struct {
	ID        string        `xml:"id,attr"`
	Severity  string        `xml:"severity,attr"`
	Msg       string        `xml:"msg,attr"`
	Verbose   string        `xml:"verbose,attr"`
	Locations []CppLocation `xml:"location"`
}

type CppLocation struct {
	File   string `xml:"file,attr"`
	Line   int    `xml:"line,attr"`
	Column int    `xml:"column,attr"`
	Info   string `xml:"info,attr"`
}
