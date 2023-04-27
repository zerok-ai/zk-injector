package inject

import (
	"fmt"
	"strings"

	"zerok-injector/pkg/common"
	utils "zerok-injector/pkg/utils"
)

var java_runtime = "java"
var java_agent_options = []string{"-javaagent:/opt/zerok/opentelemetry-javaagent.jar", "-Dotel.javaagent.extensions=/opt/zerok/zk-otel-extension.jar", "-Dotel.traces.exporter=logging"}

func transformCommandAndArgsK8s(command string, runtime common.ProgrammingLanguage) ([]string, error) {
	if runtime == common.JavaProgrammingLanguage {
		transformedCommand, err := transformForJavaRuntime(command, runtime)
		if err != nil {
			fmt.Printf("Error while transforming command %v, err %v", command, err)
			return []string{}, fmt.Errorf("error while transforming the command")
		}
		return transformedCommand, nil
	}
	return []string{}, fmt.Errorf("unkown programming language")
}

// This function transforms the cmd and args if found to be of java runtime.
// It returns the instrumented cmd and args along with bool which specifies whether
// if the instrumentation is done.
func transformForJavaRuntime(command string, runtime common.ProgrammingLanguage) ([]string, error) {
	commandArr := strings.Split(command, " ")
	index := utils.FindString(commandArr, java_runtime)
	if index >= 0 {
		commandArr = utils.AppendArray(commandArr, java_agent_options, index+1)
	} else {
		return []string{}, fmt.Errorf("not found java in command")
	}
	return commandArr, nil
}
