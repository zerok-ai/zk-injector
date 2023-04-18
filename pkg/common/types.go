package common

type ProcessDetails struct {
	ProcessID   int
	ExeName     string
	CmdLine     string
	Runtime     ProgrammingLanguage
	ProcessName string
}

type ProgrammingLanguage string

const (
	JavaProgrammingLanguage       ProgrammingLanguage = "java"
	PythonProgrammingLanguage     ProgrammingLanguage = "python"
	GoProgrammingLanguage         ProgrammingLanguage = "go"
	DotNetProgrammingLanguage     ProgrammingLanguage = "dotnet"
	JavascriptProgrammingLanguage ProgrammingLanguage = "javascript"
	UknownLanguage                ProgrammingLanguage = "unknown"
)

type ContainerRuntime struct {
	PodUID        string           `json:"uid"`
	ContainerName string           `json:"cont"`
	Image         string           `json:"image"`
	Process       []ProcessDetails `json:"process"`
}

type RuntimeSyncRequest struct {
	RuntimeDetails []ContainerRuntime `json:"details"`
}
