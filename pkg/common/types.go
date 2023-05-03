package common

type ProcessDetails struct {
	ProcessID   int                 `json:"pid"`
	ExeName     string              `json:"exe"`
	CmdLine     string              `json:"cmd"`
	Runtime     ProgrammingLanguage `json:"runtime"`
	ProcessName string              `json:"pname"`
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

// Ques: Are PodUID and ContainerName needed?
type ContainerRuntime struct {
	PodUID        string           `json:"uid"`
	ContainerName string           `json:"cont"`
	Image         string           `json:"image"`
	ImageID       string           `json:"imageId"`
	Process       []ProcessDetails `json:"process"`
}

type RuntimeSyncRequest struct {
	RuntimeDetails []ContainerRuntime `json:"details"`
}
