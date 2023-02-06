package zkclient

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	dockerConfigKey string = ".dockerconfigjson"
	authsKey        string = "auths"
)

func getImagePullSecret(pod *corev1.Pod) []string {
	imagePullSecrets := &pod.Spec.ImagePullSecrets

	var secrets []string = []string{}

	for _, imagePullSecret := range *imagePullSecrets {
		secrets = append(secrets, imagePullSecret.Name)
	}
	return secrets
}

func GetK8sClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
