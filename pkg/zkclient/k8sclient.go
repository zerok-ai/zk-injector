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
	k8sClient := GetK8sClient()
	deploymentsClient := k8sClient.AppsV1().Deployments(namespace)
	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"zk-operator/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))
	//TODO: Do we need to add any filter based on label for getting the deployments? Something about prev orchestration status.
	_, err := deploymentsClient.Patch(context.TODO(), deployment, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		fmt.Printf("Error caught while restarting deployment %v.\n", err)
		return err
	}
	return nil
}

// TODO: Confirm this with shivam.
func RestartAllDeplomentsInNamespace(namespace string) {
	k8sClient := GetK8sClient()
	deployments, err := k8sClient.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting deployments: %v\n", err)
		return
	}

	for _, deployment := range deployments.Items {
		fmt.Printf("Restarting Deployment: %s\n", deployment.ObjectMeta.Name)
		RestartDeployment(namespace, deployment.ObjectMeta.Name)
	}
}

func LabelPod(pod *corev1.Pod, path string, value string) {
	k8sClient := GetK8sClient().CoreV1()
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  path,
		Value: value,
	}}
	payloadBytes, _ := json.Marshal(payload)
	_, updateErr := k8sClient.Pods(pod.GetNamespace()).Patch(context.Background(), pod.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if updateErr == nil {
		logMessage := fmt.Sprintf("Pod %s labeled successfully for Path %s and Value %s.", pod.GetName(), path, value)
		fmt.Println(logMessage)
	} else {
		fmt.Println(updateErr)
	}
}

func getPodsWithSelector(selector string) *corev1.PodList {
	clientset := GetK8sClient()
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	pods, _ := clientset.CoreV1().Pods("").List(context.Background(), listOptions)
	return pods
}

func GetPodsMatchingLabels(labelsMap map[string]string, namespace string) (*corev1.PodList, error) {
	clientset := GetK8sClient()
	labelSet := labels.Set(labelsMap)
	listOptions := metav1.ListOptions{
		LabelSelector: labelSet.AsSelector().String(),
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	return pods, err
}

func GetPodsWithLabel(labelKey, labelValue string) *corev1.PodList {
	return getPodsWithSelector(labelKey + "=" + labelValue)
}

func GetPodsWithoutLabel(labelKey string) *corev1.PodList {
	return getPodsWithSelector("!" + labelKey)
}

func CreatePod(pod *corev1.Pod) (*corev1.Pod, error) {
	clientset := GetK8sClient()
	pod, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	return pod, err
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
