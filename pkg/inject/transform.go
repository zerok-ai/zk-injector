package inject

import utils "github.com/zerok-ai/zerok-injector/pkg/utils"

var java_runtime = "java"
var java_agent_options = []string{"-javaagent:/opt/zerok/opentelemetry-javaagent.jar", "-Dotel.javaagent.extensions=/opt/zerok/zk-otel-extension.jar", "-Dotel.traces.exporter=logging"}

func transformCommandAndArgsK8s(command, args []string) ([]string, []string) {
	command, args, _ = transformForJavaRuntime(command, args)
	return command, args
}

// This function transforms the cmd and args if found to be of java runtime.
// It returns the instrumented cmd and args along with bool which specifies whether
// if the instrumentation is done.
func transformForJavaRuntime(command, args []string) ([]string, []string, bool) {
	instrumented := false
	index := utils.FindString(command, java_runtime)
	if index >= 0 {
		command = utils.AppendArray(command, java_agent_options, index+1)
		instrumented = true
	} else {
		index = utils.FindString(args, java_runtime)
		if index >= 0 {
			args = utils.AppendArray(args, java_agent_options, index+1)
		}
		instrumented = true
	}
	return command, args, instrumented
}
