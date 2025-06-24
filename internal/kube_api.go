package internal

import (
	"context"
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServerSettingsStorage interface {
	ReadServers() ([]Server, error)
}

type KubeApi struct {
	clientset *kubernetes.Clientset
	namespace string
}

func NewKubeApi(namespace string) *KubeApi {
	home := os.Getenv("HOME")
	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return &KubeApi{clientset: clientset, namespace: namespace}
}

func (api *KubeApi) ReadServers() ([]Server, error) {
	ingresses, err := api.clientset.NetworkingV1().Ingresses(api.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	secrets, err := api.clientset.CoreV1().Secrets(api.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var servers []Server
	for _, ingress := range ingresses.Items {
		annotations := ingress.Annotations
		domain, hasDomain := annotations["external-proxy/domain"]
		serverSnippets, hasServerSnippets := annotations["external-proxy/server-snippets"]
		if !hasDomain || !hasServerSnippets {
			continue
		}

		server := Server{
			Domain:   domain,
			Snippet: string(serverSnippets),
		}

		for _, secret := range secrets.Items {
			if secret.Type == "kubernetes.io/tls" {
				secretDomain, ok := secret.Annotations["external-proxy/domain"]
				if !ok || secretDomain != domain {
					continue
				}
				crt, crtOk := secret.Data["tls.crt"]
				key, keyOk := secret.Data["tls.key"]
				if !crtOk || !keyOk {
					fmt.Printf("Secret %s missing tls.crt or tls.key\n", secretDomain)
					continue
				}

				cert := Cert{
					Certificate: string(crt),
					PrivateKey:  string(key),
				}

				server.Cert = cert
			}
		}

		servers = append(servers, server)
	}

	return servers, nil
}
