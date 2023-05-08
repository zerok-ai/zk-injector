package common

const (
	ZkOrchKey          = "zk-status"
	ZkOrchPath         = "/metadata/labels/" + ZkOrchKey
	ZkOrchOrchestrated = "orchestrated"
	ZkOrchProcessed    = "processed"

	JavalToolOptions = "JAVA_TOOL_OPTIONS"
	OtelArgument     = " -javaagent:/opt/zerok/opentelemetry-javaagent.jar -Dotel.traces.exporter=jaeger -Dotel.exporter.jaeger.endpoint=simplest-collector.observability.svc.cluster.local:14250"

	ZkInjectionKey   = "zk-injection"
	ZkInjectionValue = "enabled"
)
