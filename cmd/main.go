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
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"reflect"
	"time"

	m "github.com/zerok-ai/zerok-injector/pkg/inject"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	webhookName        = "zk-webhook"
	webhookPath        = "/zk-injector"
	webhookNamespace   = "zk-injector"
	webhookServiceName = "zk-injector"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request received at Path %q\n", r.URL.Path)
}

func injectRequestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request recevied.\n")
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		errorResponse(err, w)
		return
	}

	modified, err := m.Inject(body, true)

	if err != nil {
		errorResponse(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(modified)

	r.Body.Close()
}

func errorResponse(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
}

func main() {
	dnsNames := []string{
		webhookServiceName,
		webhookServiceName + "." + webhookNamespace,
		webhookServiceName + "." + webhookNamespace + ".svc",
	}
	commonName := webhookServiceName + "." + webhookNamespace + ".svc"

	org := "zerok"
	caPEM, certPEM, certKeyPEM, err := generateCert([]string{org}, dnsNames, commonName)
	if err != nil {
		fmt.Printf("Failed to generate ca and certificate key pair: %v.\n", err)
	}

	pair, err := tls.X509KeyPair(certPEM.Bytes(), certKeyPEM.Bytes())
	if err != nil {
		fmt.Printf("Failed to load certificate key pair: %v.\n", err)
	}

	// create or update the mutatingwebhookconfiguration
	err = createOrUpdateMutatingWebhookConfiguration(caPEM, webhookServiceName, webhookNamespace)
	if err != nil {
		fmt.Printf("Failed to create or update the mutating webhook configuration: %v\n", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", defaultHandler)
	mux.HandleFunc("/zk-injector", injectRequestHandler)

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		TLSConfig:      &tls.Config{Certificates: []tls.Certificate{pair}},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
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
	fail := admissionregistrationv1.Fail
	sideEffect := admissionregistrationv1.SideEffectClassNone
	mutatingWebhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name:                    "zk-webhook.zerok.ai",
			AdmissionReviewVersions: []string{"v1"},
			SideEffects:             &sideEffect,
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caPEM.Bytes(), // self-generated CA for the webhook
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
			FailurePolicy: &fail,
		}},
	}

	foundWebhookConfig, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Get(context.TODO(), webhookName, metav1.GetOptions{})
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
	} else {
		// there is an existing mutatingWebhookConfiguration
		if len(foundWebhookConfig.Webhooks) != len(mutatingWebhookConfig.Webhooks) ||
			!(foundWebhookConfig.Webhooks[0].Name == mutatingWebhookConfig.Webhooks[0].Name &&
				reflect.DeepEqual(foundWebhookConfig.Webhooks[0].AdmissionReviewVersions, mutatingWebhookConfig.Webhooks[0].AdmissionReviewVersions) &&
				reflect.DeepEqual(foundWebhookConfig.Webhooks[0].SideEffects, mutatingWebhookConfig.Webhooks[0].SideEffects) &&
				reflect.DeepEqual(foundWebhookConfig.Webhooks[0].FailurePolicy, mutatingWebhookConfig.Webhooks[0].FailurePolicy) &&
				reflect.DeepEqual(foundWebhookConfig.Webhooks[0].Rules, mutatingWebhookConfig.Webhooks[0].Rules) &&
				reflect.DeepEqual(foundWebhookConfig.Webhooks[0].NamespaceSelector, mutatingWebhookConfig.Webhooks[0].NamespaceSelector) &&
				reflect.DeepEqual(foundWebhookConfig.Webhooks[0].ClientConfig.CABundle, mutatingWebhookConfig.Webhooks[0].ClientConfig.CABundle) &&
				reflect.DeepEqual(foundWebhookConfig.Webhooks[0].ClientConfig.Service, mutatingWebhookConfig.Webhooks[0].ClientConfig.Service)) {
			mutatingWebhookConfig.ObjectMeta.ResourceVersion = foundWebhookConfig.ObjectMeta.ResourceVersion
			if _, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Update(context.TODO(), mutatingWebhookConfig, metav1.UpdateOptions{}); err != nil {
				fmt.Printf("Failed to update the mutatingwebhookconfiguration: %s", webhookName)
				return err
			}
			fmt.Printf("Updated the mutatingwebhookconfiguration: %s\n", webhookName)
		}
		fmt.Printf("The mutatingwebhookconfiguration: %s already exists and has no change\n", webhookName)
	}

	return nil
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

	newCertPEM, newPrivateKeyPEM, err := getClientCertPEM(orgs, dnsNames, commonName, ca, caPrivateKey)

	if err != nil {
		return nil, nil, nil, err
	}

	return caPEM, newCertPEM, newPrivateKeyPEM, nil
}

func getClientCertPEM(orgs, dnsNames []string, commonName string, parentCa *x509.Certificate, parentPrivateKey *rsa.PrivateKey) (*bytes.Buffer, *bytes.Buffer, error) {
	newCert := &x509.Certificate{
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

	newPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	newCertBytes, err := x509.CreateCertificate(rand.Reader, newCert, parentCa, &newPrivateKey.PublicKey, parentPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	newCertPEM := new(bytes.Buffer)
	_ = pem.Encode(newCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: newCertBytes,
	})

	newPrivateKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(newPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(newPrivateKey),
	})

	return newCertPEM, newPrivateKeyPEM, nil
}
