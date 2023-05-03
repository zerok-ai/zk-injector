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

agent_options="-javaagent:/opt/zerok/opentelemetry-javaagent.jar -Dotel.traces.exporter=jaeger -Dotel.exporter.jaeger.endpoint=simplest-collector.observability.svc.cluster.local:14250"
javaString="java $agent_options"

firstCommand=$1
finalCommand=${firstCommand/java/$javaString}



echo $finalCommand

# $final_cmd

eval "$finalCommand"
