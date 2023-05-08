package zkclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	//TODO: Do we need to add any filter based on label for getting the deployments? Something about prev orchestration status.
	_, err = deploymentsClient.Patch(context.TODO(), deployment, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		fmt.Printf("Error caught while restarting deployment %v.\n", err)
		return err
	}
	return nil
}

// TODO: Confirm this with shivam.
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
	//TODO: Do we also need to restart other workloads like statefulsets?
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

func getPodsWithSelector(selector string) (*corev1.PodList, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return nil, err
	}
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	pods, _ := clientset.CoreV1().Pods("").List(context.Background(), listOptions)
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

func GetPodsWithLabel(labelKey, labelValue string) (*corev1.PodList, error) {
	pods, err := getPodsWithSelector(labelKey + "=" + labelValue)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func GetPodsWithoutLabel(labelKey string) (*corev1.PodList, error) {
	pods, err := getPodsWithSelector("!" + labelKey)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func GetK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
