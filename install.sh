make docker push
kubectl create namespace zk-injector
kubectl label namespace default zk-injection=enabled
kubectl apply -f deploy/webhook.yaml -n zk-injector