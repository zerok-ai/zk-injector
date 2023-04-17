package detector

import (
	"context"
	"fmt"
	"strings"

	"github.com/zerok-ai/zerok-injector/common"
	"github.com/zerok-ai/zerok-injector/pkg/zkclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	istioAnnotationKey     = "sidecar.istio.io/inject"
	istioAnnotationValue   = "false"
	linkerdAnnotationKey   = "linkerd.io/inject"
	linkerdAnnotationValue = "disabled"
)

func detectLanguage(ctx context.Context, namespace string, labels map[string]string) error {
	targetPod, err := choosePods(ctx, labels, namespace)
	if err != nil {
		return err
	}

	langDetectionPod, err := createLangDetectionPod(targetPod)
	if err != nil {
		return err
	}

	_, err = zkclient.CreatePod(langDetectionPod)
	return err
}

func choosePods(ctx context.Context, labels map[string]string, namespace string) (*corev1.Pod, error) {
	podList, err := zkclient.GetPodsMatchingLabels(labels, namespace)
	if err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		return nil, common.PodsNotFoundErr
	}

	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodRunning {
			return &pod, nil
		}
	}

	return nil, common.PodsNotFoundErr
}

func createLangDetectionPod(targetPod *corev1.Pod) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-lang-detection-", targetPod.Name),
			Namespace:    targetPod.Namespace,
			Annotations: map[string]string{
				common.LangDetectionContainerAnnotationKey: "true",
				istioAnnotationKey:                         istioAnnotationValue,
				linkerdAnnotationKey:                       linkerdAnnotationValue,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "lang-detector",
					Image: fmt.Sprintf("%s:%s", common.LangDetectorImage, common.LangDetectorTag),
					Args: []string{
						fmt.Sprintf("--pod-uid=%s", targetPod.UID),
						fmt.Sprintf("--container-names=%s", strings.Join(getContainerNames(targetPod), ",")),
					},
					TerminationMessagePath: "/dev/detection-result",
					SecurityContext: &corev1.SecurityContext{
						Capabilities: &corev1.Capabilities{
							Add: []corev1.Capability{"SYS_PTRACE"},
						},
					},
				},
			},
			RestartPolicy: "Never",
			NodeName:      targetPod.Spec.NodeName,
			HostPID:       true,
		},
	}

	// err := ctrl.SetControllerReference(instrumentedApp, pod, r.Scheme)
	// if err != nil {
	// 	return nil, err
	// }

	return pod, nil
}

func getContainerNames(pod *corev1.Pod) []string {
	var result []string
	for _, c := range pod.Spec.Containers {
		if !skipContainer(c.Name) {
			result = append(result, c.Name)
		}
	}

	return result
}

func skipContainer(name string) bool {
	return name == "istio-proxy" || name == "linkerd-proxy"
}
