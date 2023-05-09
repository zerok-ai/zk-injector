# zerok-injector
Zerok injector in responsible for instrumenting the pods coming up in the cluster. It achieves this by adding mutatingadmissionwebhook in the cluster. 

# Pre-requisites
It needs redis to be up and running as it uses redis to read the language data for images in the cluster. This data will be populated by the zerok-deamonset pod. The injector pod will sync the data from redis based on time interval specified in the config file.

### Quickstart

1. To run the injector on your local machine. This script reads config from internal/config-local.yaml file.

```
./scripts/runLocal.sh
```

### Installing on the cluster
1. Build and push the injector image. This script will create the image using config from internal/config.yaml file.

```
./build.sh
```

2. Applying the injector on the cluster.
	
```
kubectl apply -f k8s.yaml
```