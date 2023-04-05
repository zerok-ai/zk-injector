echo "Agent Injected successfully."
echo
echo "OS details:"
cat /etc/os-release
echo 

echo 
echo "PWD:"
pwd
echo "ls:zerok"
ls /opt/zerok/
echo "ls:"
ls

agent_options="-javaagent:/opt/zerok/opentelemetry-javaagent.jar -Dotel.javaagent.extensions=/opt/zerok/zk-otel-extension.jar -Dotel.traces.exporter=zipkin -Dotel.exporter.zipkin.endpoint=http://zipkin.default.svc.cluster.local:9411/api/v2/spans"
javaString="java $agent_options"

firstCommand=$1
finalCommand=${firstCommand/java/$javaString}



echo $finalCommand

# $final_cmd

eval "$finalCommand"
