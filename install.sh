make docker-build docker-push
kubectl create namespace zerok
kubectl label namespace zerok zk-injection=enabled
kubectl apply -f deploy/webhook.yaml -n zerok