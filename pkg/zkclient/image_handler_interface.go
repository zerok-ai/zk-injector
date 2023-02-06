package zkclient

import corev1 "k8s.io/api/core/v1"

type ImageHandlerInterface interface {
	pullImage(image string, pod *corev1.Pod) error
	GetCommandFromImage(image string, pod *corev1.Pod, tracker *ImageDownloadTracker) ([]string, error)
}
