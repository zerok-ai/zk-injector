package zkclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"os"

	"zerok-injector/pkg/common"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func RestartDeployment(namespace string, deployment string) error {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}
	deploymentsClient := k8sClient.AppsV1().Deployments(namespace)
	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"zk-operator/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))
	_, err = deploymentsClient.Patch(context.TODO(), deployment, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		fmt.Printf("Error caught while restarting deployment %v.\n", err)
		return err
	}
	return nil
}

func RestartAllDeplomentsInNamespace(namespace string) error {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}
	deployments, err := k8sClient.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting deployments: %v\n", err)
		return err
	}

	for _, deployment := range deployments.Items {
		fmt.Printf("Restarting Deployment: %s\n", deployment.ObjectMeta.Name)
		RestartDeployment(namespace, deployment.ObjectMeta.Name)
	}
	return nil
}

func LabelPod(pod *corev1.Pod, path string, value string) error {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  path,
		Value: value,
	}}
	payloadBytes, _ := json.Marshal(payload)
	_, updateErr := k8sClient.CoreV1().Pods(pod.GetNamespace()).Patch(context.Background(), pod.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if updateErr == nil {
		logMessage := fmt.Sprintf("Pod %s labeled successfully for Path %s and Value %s.", pod.GetName(), path, value)
		fmt.Println(logMessage)
		return updateErr
	} else {
		fmt.Println(updateErr)
	}
	return nil
}

func getPodsWithSelector(selector string, namespace string) (*corev1.PodList, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return nil, err
	}
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	pods, _ := clientset.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	return pods, nil
}

func GetPodsMatchingLabels(labelsMap map[string]string, namespace string) (*corev1.PodList, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return nil, err
	}
	labelSet := labels.Set(labelsMap)
	listOptions := metav1.ListOptions{
		LabelSelector: labelSet.AsSelector().String(),
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	return pods, err
}

func GetPodsWithLabel(labelKey, labelValue, namespace string) (*corev1.PodList, error) {
	pods, err := getPodsWithSelector(labelKey+"="+labelValue, namespace)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func GetPodsWithoutLabel(labelKey string, namespace string) (*corev1.PodList, error) {
	pods, err := getPodsWithSelector("!"+labelKey, namespace)
	if err != nil {
		fmt.Printf("Error while getting pods without label %v.\n", err)
		return nil, err
	}
	return pods, nil
}

func GetAllNonOrchestratedPods() ([]corev1.Pod, error) {
	allPodsList := []corev1.Pod{}
	clientset, err := GetK8sClient()
	if err != nil {
		fmt.Printf(" Error while getting client.")
		return nil, err
	}
	selector := common.ZkInjectionKey + "=" + common.ZkInjectionValue
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), listOptions)
	if err != nil {
		fmt.Printf("Error caught while getting list of namespacese %v.\n", err)
	}
	for _, namespace := range namespaces.Items {
		fmt.Printf("Checking for namespace %v.\n", namespace)
		pods, err := GetNotOrchestratedPods(namespace.ObjectMeta.Name)
		if err != nil {
			err = fmt.Errorf("error getting non orchestrated pods from namespace %v", namespace)
			return nil, err
		}
		allPodsList = append(allPodsList, pods.Items...)
	}
	return allPodsList, nil
}

func GetNotOrchestratedPods(namespace string) (*corev1.PodList, error) {
	pods, err := GetPodsWithoutLabel(common.ZkOrchKey, namespace)
	return pods, err
}

func GetK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// If incluster config failes, reading from kubeconfig.
		// However, this is not connecting to gcp clusters. Only working for kind now(probably minikube also).
		kubeconfig := os.Getenv("KUBECONFIG")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes config: %v", err)
		}
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
