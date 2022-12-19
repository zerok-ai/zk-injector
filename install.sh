make docker push
kubectl create namespace zk-injector
kubectl apply -f deploy/webhook.yaml -n zk-injector