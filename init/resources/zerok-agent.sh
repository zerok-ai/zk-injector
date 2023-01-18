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
echo "ls:"
ls

if [["$0" == *"java"*]]; then
   echo "java program running."
fi 
