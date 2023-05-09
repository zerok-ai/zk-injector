# Zerok-injector
Zerok injector in responsible for instrumenting the pods coming up in the cluster. It achieves this by adding a mutatingadmissionwebhook in the cluster.

# Pre-requisites
It needs Redis to be up and running as it uses Redis to read language data for images. This data will be populated by the zerok-deamonset pod. The injector pod will sync the data from Redis based on the time interval specified in the config file.

### Quickstart

1. To run the injector on your local machine, run the below command. This script reads the config from the internal/config-local.yaml file.

```
./scripts/runLocal.sh
```

### Installing on the cluster
1. To build and push the injector image, run the below command. This script creates the image using config from the internal/config.yaml file.

```
./build.sh
```

2. Applying the injector on the cluster.
	
```
kubectl apply -f k8s.yaml
```

### Testing
To test the webhook request coming from the control plane, you can test using the Webhook Request present in the zk-injector postman collection. 