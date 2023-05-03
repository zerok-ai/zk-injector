echo "Agent Injected successfully."
echo
echo "OS details:"
cat /etc/os-release
echo 

# mv node_modules/express node_modules/express_real
# mkdir node_modules/express
# cp -r /opt/zerok/zerok-node-module/* node_modules/express/

# echo "---- npm install ----"
# npm install @opentelemetry/sdk-node @opentelemetry/api
# npm install @opentelemetry/auto-instrumentations-node

echo 
echo "PWD:"
pwd
echo "ls:zerok"
ls /opt/zerok/
echo "ls:"
ls

echo "--------------------"

final_cmd=""
agent_options="-javaagent:/opt/zerok/opentelemetry-javaagent.jar -Dotel.service.name=app -Dotel.javaagent.extensions=/opt/zerok/zk-otel-extension.jar -Dotel.traces.exporter=jaeger -Dotel.exporter.jaeger.endpoint=http://simplest-collector.observability.svc.cluster.local:14250"

for var in "$@"
do
    final_cmd="$final_cmd $var"
    if [ "$var" = "java" ]
    then
        final_cmd="$final_cmd $agent_options"
    fi
done

echo $final_cmd

eval $final_cmd