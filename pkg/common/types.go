package common

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
	Image     string   `json:"image"`
	ImageID   string   `json:"imageId"`
	Languages []string `json:"language"`
}
