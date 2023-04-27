package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"
	"zerok-injector/internal/config"

	"github.com/ilyakaznacheev/cleanenv"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"zerok-injector/pkg/inject"
	"zerok-injector/pkg/server"
	"zerok-injector/pkg/storage"
)

var (
	webhookName        = "zerok-webhook"
	webhookPath        = "/zk-injector"
	webhookNamespace   = "zerok-injector"
	webhookServiceName = "zerok-injector"
)

type HttpApiHandler struct {
	injector *inject.Injector
}

func (h *HttpApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	fmt.Printf("Got a request from webhook")

	if err != nil {
		errorResponse(err, w)
		return
	}

	response, err := h.injector.Inject(body)

	if err != nil {
		fmt.Printf("Error while injecting zk agent %v\n", err)
	}

	// Sending http status as OK, even when injection failed to not disturb the pods in cluster.
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	r.Body.Close()
}

func errorResponse(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
}

func main() {

	var cfg config.ZkInjectorConfig
	// path := "/opt/zk-auth-configmap.yaml"
	args := ProcessArgs(&cfg)

	// read configuration from the file and environment variables
	log.Println("args.ConfigPath==", args.ConfigPath)
	if err := cleanenv.ReadConfig(args.ConfigPath, &cfg); err != nil {
		//zklogger.Error(LOG_TAG, err)
		log.Println(err)
		//os.Exit(2)
	}

	// initialize certificates
	caPEM, cert, err := initializeKeysAndCertificates()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// start mutating webhook
	err = createOrUpdateMutatingWebhookConfiguration(caPEM, webhookServiceName, webhookNamespace)
	if err != nil {
		msg := fmt.Sprintf("Failed to create or update the mutating webhook configuration: %v\n", err)
		fmt.Println(msg)
		panic(msg)
	}

	runtimeMap := &storage.ImageRuntimeHandler{ImageRuntimeMap: sync.Map{}}

	// start data collector
	go server.StartServer(runtimeMap)

	// start webhook server
	startWebHookServer(cert, runtimeMap)
}

func initializeKeysAndCertificates() (*bytes.Buffer, tls.Certificate, error) {

	dnsNames := []string{
		webhookServiceName,
		webhookServiceName + "." + webhookNamespace,
		webhookServiceName + "." + webhookNamespace + ".svc",
	}
	commonName := webhookServiceName + "." + webhookNamespace + ".svc"

	// get CA PEM
	org := "zerok"
	caPEM, serverCertPEM, serverCertKeyPEM, err := generateCert([]string{org}, dnsNames, commonName)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate certificate: %v.\n", err)
		return nil, tls.Certificate{}, fmt.Errorf(msg)
	}

	// get CA certificate
	serverPair, err := tls.X509KeyPair(serverCertPEM.Bytes(), serverCertKeyPEM.Bytes())
	if err != nil {
		msg := fmt.Sprintf("Failed to load server certificate key pair: %v.\n", err)
		return nil, tls.Certificate{}, fmt.Errorf(msg)
	}

	return caPEM, serverPair, nil
}

func startWebHookServer(serverPair tls.Certificate, runtimeMap *storage.ImageRuntimeHandler) {

	injectHandler := &HttpApiHandler{
		injector: &inject.Injector{ImageRuntimeHandler: runtimeMap},
	}

	mux := http.NewServeMux()
	mux.Handle(webhookPath, injectHandler)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		TLSConfig:      &tls.Config{Certificates: []tls.Certificate{serverPair}},
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServeTLS("", "")
}

func createOrUpdateMutatingWebhookConfiguration(caPEM *bytes.Buffer, webhookService, webhookNamespace string) error {

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	mutatingWebhookConfigV1Client := clientset.AdmissionregistrationV1()

	fmt.Printf("Creating or updating the mutatingwebhookconfiguration\n")
	ignore := admissionregistrationv1.Ignore
	sideEffect := admissionregistrationv1.SideEffectClassNone
	mutatingWebhookConfig := createMutatingWebhook(sideEffect, caPEM, webhookService, webhookNamespace, ignore)

	existingWebhookConfig, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Get(context.TODO(), webhookName, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		if _, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Create(context.TODO(), mutatingWebhookConfig, metav1.CreateOptions{}); err != nil {
			fmt.Printf("Failed to create the mutatingwebhookconfiguration: %s\n", webhookName)
			return err
		}
		fmt.Printf("Created mutatingwebhookconfiguration: %s\n", webhookName)
	} else if err != nil {
		fmt.Printf("Failed to check the mutatingwebhookconfiguration: %s\n", webhookName)
		fmt.Printf("The error is %v\n", err.Error())
		return err
	} else if !areWebHooksSame(existingWebhookConfig, mutatingWebhookConfig) {
		mutatingWebhookConfig.ObjectMeta.ResourceVersion = existingWebhookConfig.ObjectMeta.ResourceVersion
		if _, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Update(context.TODO(), mutatingWebhookConfig, metav1.UpdateOptions{}); err != nil {
			fmt.Printf("Failed to update the mutatingwebhookconfiguration: %s", webhookName)
			return err
		}
		fmt.Printf("Updated the mutatingwebhookconfiguration: %s\n", webhookName)
	} else {
		fmt.Printf("The mutatingwebhookconfiguration: %s already exists and has no change\n", webhookName)
	}

	return nil
}

func createMutatingWebhook(sideEffect admissionregistrationv1.SideEffectClass, caPEM *bytes.Buffer, webhookService string, webhookNamespace string, ignore admissionregistrationv1.FailurePolicyType) *admissionregistrationv1.MutatingWebhookConfiguration {
	timeOut := int32(30)
	mutatingWebhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name:                    "zk-webhook.zerok.ai",
			AdmissionReviewVersions: []string{"v1"},
			SideEffects:             &sideEffect,
			TimeoutSeconds:          &timeOut,
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caPEM.Bytes(),
				Service: &admissionregistrationv1.ServiceReference{
					Name:      webhookService,
					Namespace: webhookNamespace,
					Path:      &webhookPath,
				},
			},
			Rules: []admissionregistrationv1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1.OperationType{
						admissionregistrationv1.Create,
						admissionregistrationv1.Update,
					},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
					},
				},
			},
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"zk-injection": "enabled",
				},
			},
			FailurePolicy: &ignore,
		}},
	}
	return mutatingWebhookConfig
}

func areWebHooksSame(foundWebhookConfig *admissionregistrationv1.MutatingWebhookConfiguration, mutatingWebhookConfig *admissionregistrationv1.MutatingWebhookConfiguration) bool {
	if len(foundWebhookConfig.Webhooks) != len(mutatingWebhookConfig.Webhooks) {
		return false
	}
	len := len(foundWebhookConfig.Webhooks)
	for i := 0; i < len; i++ {
		equal := foundWebhookConfig.Webhooks[i].Name == mutatingWebhookConfig.Webhooks[i].Name &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].AdmissionReviewVersions, mutatingWebhookConfig.Webhooks[i].AdmissionReviewVersions) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].SideEffects, mutatingWebhookConfig.Webhooks[i].SideEffects) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].FailurePolicy, mutatingWebhookConfig.Webhooks[i].FailurePolicy) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].Rules, mutatingWebhookConfig.Webhooks[i].Rules) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].NamespaceSelector, mutatingWebhookConfig.Webhooks[i].NamespaceSelector) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].ClientConfig.CABundle, mutatingWebhookConfig.Webhooks[i].ClientConfig.CABundle) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].ClientConfig.Service, mutatingWebhookConfig.Webhooks[i].ClientConfig.Service)
		if !equal {
			return false
		}
	}
	return true
}

func generateCert(orgs, dnsNames []string, commonName string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(int64(time.Now().Day())),
		Subject:               pkix.Name{Organization: orgs},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	caPEM := new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	serverCertPEM, serverPrivateKeyPEM, err := getServerCertPEM(orgs, dnsNames, commonName, ca, caPrivateKey)

	if err != nil {
		return nil, nil, nil, err
	}

	return caPEM, serverCertPEM, serverPrivateKeyPEM, nil
}

func getServerCertPEM(orgs, dnsNames []string, commonName string, parentCa *x509.Certificate, parentPrivateKey *rsa.PrivateKey) (*bytes.Buffer, *bytes.Buffer, error) {
	serverCert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1024),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: orgs,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverCert, parentCa, &serverPrivateKey.PublicKey, parentPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	serverCertPEM := new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivateKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(serverPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})

	return serverCertPEM, serverPrivateKeyPEM, nil
}

// Args command-line parameters
type Args struct {
	ConfigPath string
}

// ProcessArgs processes and handles CLI arguments
func ProcessArgs(cfg interface{}) Args {
	var a Args

	f := flag.NewFlagSet("Example server", 1)
	f.StringVar(&a.ConfigPath, "c", "config.yaml", "Path to configuration file")

	fu := f.Usage
	f.Usage = func() {
		fu()
		envHelp, _ := cleanenv.GetDescription(cfg, nil)
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), envHelp)
	}

	f.Parse(os.Args[1:])
	return a
}
