package zkclient

type ImageHandlerInterface interface {
	pullImage(image string) error
	GetCommandFromImage(image string, tracker *ImageDownloadHandler, handler *ImageHandlerInterface) ([]string, error)
}
