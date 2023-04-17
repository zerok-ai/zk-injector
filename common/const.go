package common

import "errors"

const (
	LangDetectionContainerAnnotationKey = "keyval.dev/lang-detection-pod"
	LangDetectorContainer               = "keyval/lang-detector"
	LangDetectionEnvVar                 = "LANG_DETECTION_VERSION"
	DefaultLangDetectionVersion         = "v0.0.249"
	CurrentNamespaceEnvVar              = "CURRENT_NS"
	DefaultNamespace                    = "odigos-system"
	DefaultOdigosConfigurationName      = "odigos-config"
	OTLPPort                            = 4317
	OTLPHttpPort                        = 4318
	LangDetectorImage                   = "rajeevzerok/zerok-langdetector"
	LangDetectorTag                     = "test"
)

var (
	PodsNotFoundErr = errors.New("could not find a ready pod")
)
