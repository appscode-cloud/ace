/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package installer_precheck

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"go.bytebuilders.dev/installer/apis/installer/v1alpha1"
	verifier "go.bytebuilders.dev/license-verifier"
	"go.bytebuilders.dev/license-verifier/info"
	configapi "go.bytebuilders.dev/resource-model/apis/config/v1alpha1"

	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	modstring "gomodules.xyz/x/strings"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"kmodules.xyz/go-containerregistry/authn"
	"sigs.k8s.io/yaml"
)

type AceValidateOptions struct {
	Licenses       map[string]string        `json:"licenses,omitempty"`
	Registry       v1alpha1.RegistrySpec    `json:"registry,omitempty"`
	SelfManagement configapi.SelfManagement `json:"selfManagement,omitempty"`
}

// CheckStatus holds the overall check status and logs.
type CheckStatus struct {
	AllOk bool
	Logs  []string
}

// LogError records an error message and updates the status.
func (s *CheckStatus) LogError(message string) {
	s.AllOk = false
	s.Logs = append(s.Logs, message)
	log.Println(message)
}

// CheckOptions validates the options from the specified path.
func CheckOptions(optsPath string) (bool, error) {
	data, err := os.ReadFile(optsPath)
	if err != nil {
		return false, fmt.Errorf("failed to read options file: %w", err)
	}

	var aceOptions AceValidateOptions
	if err := yaml.Unmarshal(data, &aceOptions); err != nil {
		return false, err
	}

	rc, err := rest.InClusterConfig()
	if err != nil {
		return false, err
	}

	kc, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return false, fmt.Errorf("failed to create clientset: %v", err)
	}

	status := CheckStatus{AllOk: true}

	// 1. Verify client-cert and client-key
	if aceOptions.Registry.Certs.ClientCert != "" && aceOptions.Registry.Certs.ClientKey != "" {
		if err := checkCertKeyPair(aceOptions.Registry.Certs.ClientCert, aceOptions.Registry.Certs.ClientKey); err != nil {
			status.LogError(fmt.Sprintf("failed to verify client-cert and client-key: %s", err))
		}
	}

	// 2. Check image pull secrets and registry credentials
	checkImagePullSecrets(&status, kc, aceOptions)

	// 3. Check docker registry URL
	if aceOptions.Registry.Image.Proxies.DockerHub != "" {
		checkDockerRegistry(aceOptions, &status)
	}

	// 4. Check kube-apiserver can be detected
	if aceOptions.SelfManagement.Import {
		if err := detectKubeAPIServer(aceOptions, rc); err != nil {
			status.LogError(fmt.Sprintf("failed to detect kube-apiserver: %s", err))
		}
	}

	// 5. Check licenses
	checkLicenses(aceOptions, kc, &status)

	// 6. Check disabled features
	checkDisabledFeatures(kc, aceOptions, &status)

	// 7. Check default storage class exists
	if err := checkDefaultStorageClassExists(kc); err != nil {
		status.LogError(fmt.Sprintf("failed to get default storage class: %s", err))
	}

	return status.AllOk, nil
}

// checkImagePullSecrets verifies that image pull secrets exist.
func checkImagePullSecrets(status *CheckStatus, kc *kubernetes.Clientset, aceOptions AceValidateOptions) {
	if len(aceOptions.Registry.ImagePullSecrets) > 0 {
		for _, sec := range aceOptions.Registry.ImagePullSecrets {
			_, err := kc.CoreV1().Secrets("kubedb").Get(context.Background(), sec, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					status.LogError(fmt.Sprintf("%s image pull secret not found", sec))
				} else {
					status.LogError(fmt.Sprintf("failed to get image pull secret %s: %s", sec, err))
				}
			}
		}

		// Check registry credentials
		_, err := authn.ImageWithDigest(kc, fmt.Sprintf("%s/ace-installer", aceOptions.Registry.Image.Proxies.AppsCode), &k8schain.Options{
			ImagePullSecrets: aceOptions.Registry.ImagePullSecrets,
		})
		if err != nil {
			status.LogError(fmt.Sprintf("failed to verify registry credentials: %s", err))
		}
	}
}

// checkDockerRegistry verifies the connection to the specified Docker registry.
func checkDockerRegistry(aceOptions AceValidateOptions, status *CheckStatus) {
	url := "https://" + aceOptions.Registry.Image.Proxies.DockerHub
	resp, err := http.Get(url)
	if err != nil {
		status.LogError(fmt.Sprintf("Error making request to Docker registry: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		status.LogError("No TLS connection or no peer certificates.")
	}
}

// checkLicenses verifies the licenses specified in aceOptions.
func checkLicenses(aceOptions AceValidateOptions, kc *kubernetes.Clientset, status *CheckStatus) {
	ns, err := kc.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	if err != nil {
		status.LogError(fmt.Sprintf("failed to get kube-system namespace: %s", err))
		return
	}

	ca, err := info.LoadLicenseCA()
	if err != nil {
		status.LogError(fmt.Sprintf("failed to get license CA: %s", err))
		return
	}

	caCert, err := info.ParseCertificate(ca)
	if err != nil {
		status.LogError(fmt.Sprintf("failed to parse license CA: %s", err))
		return
	}

	for pro, lic := range aceOptions.Licenses {
		decodedData, err := base64.StdEncoding.DecodeString(lic)
		if err != nil {
			status.LogError(fmt.Sprintf("failed to Base64-decode data: %v", err))
			return
		}
		if _, err := verifier.ParseLicense(verifier.ParserOptions{
			ClusterUID: string(ns.UID),
			CACert:     caCert,
			License:    decodedData,
		}); err != nil {
			status.LogError(fmt.Sprintf("failed to verify license for product %s: %s", pro, err))
		}
	}
}

// checkDisabledFeatures checks for any disabled features and their existence.
func checkDisabledFeatures(kc *kubernetes.Clientset, aceOptions AceValidateOptions, status *CheckStatus) {
	features := map[string]func(*kubernetes.Clientset, *CheckStatus){
		"kubedb":       checkKubeDBExists,
		"stash":        checkStashExists,
		"kubestash":    checkKubeStashExists,
		"cert-manager": checkCertManagerExists,
	}

	for feature, checkFunc := range features {
		if modstring.Contains(aceOptions.SelfManagement.DisableFeatures, feature) {
			checkFunc(kc, status)
		}
	}
}

// checkDeploymentExists checks if a deployment exists based on a label selector.
func checkDeploymentExists(kc *kubernetes.Clientset, labelSelector map[string]string) (bool, error) {
	selector := labels.Set(labelSelector).AsSelector().String()

	deployments, err := kc.AppsV1().Deployments(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return false, err
	}

	return len(deployments.Items) > 0, nil
}

// checkFeatureExists checks if a specific feature exists.
func checkFeatureExists(kc *kubernetes.Clientset, featureName string, labels map[string]string, status *CheckStatus) {
	found, err := checkDeploymentExists(kc, labels)
	if err != nil {
		status.LogError(fmt.Sprintf("failed to get %s deployments: %s", featureName, err))
		return
	}

	if !found {
		status.LogError(fmt.Sprintf("%s not found in this cluster", featureName))
	}
}

// checkKubeDBExists checks if KubeDB is deployed in the cluster.
func checkKubeDBExists(kc *kubernetes.Clientset, status *CheckStatus) {
	kubedbLabels := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "kubedb-provisioner",
	}
	checkFeatureExists(kc, "kubedb", kubedbLabels, status)
}

// checkStashExists checks if Stash is deployed in the cluster.
func checkStashExists(kc *kubernetes.Clientset, status *CheckStatus) {
	stashLabels := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "stash-enterprise",
	}
	checkFeatureExists(kc, "stash", stashLabels, status)
}

// checkKubeStashExists checks if KubeStash is deployed in the cluster.
func checkKubeStashExists(kc *kubernetes.Clientset, status *CheckStatus) {
	kubestashLabels := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "kubestash",
	}
	checkFeatureExists(kc, "kubestash", kubestashLabels, status)
}

// checkCertManagerExists checks if cert-manager is deployed in the cluster.
func checkCertManagerExists(kc *kubernetes.Clientset, status *CheckStatus) {
	certManagerLabels := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "cert-manager",
	}
	checkFeatureExists(kc, "cert-manager", certManagerLabels, status)
}

// checkDefaultStorageClassExists verifies that a default storage class exists in the cluster.
func checkDefaultStorageClassExists(kc *kubernetes.Clientset) error {
	storageClasses, err := kc.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list storage classes: %w", err)
	}

	for _, sc := range storageClasses.Items {
		if sc.Provisioner == "kubernetes.io/no-provisioner" {
			continue
		}
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			return nil
		}
	}
	return errors.New("no default storage class found")
}

// checkCertKeyPair verifies if the provided client certificate and key pair is valid.
func checkCertKeyPair(clientCertPath string, clientKeyPath string) error {
	clientCertPEM, err := os.ReadFile(clientCertPath)
	if err != nil {
		return fmt.Errorf("failed to read client certificate: %w", err)
	}

	clientKeyPEM, err := os.ReadFile(clientKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read client key: %w", err)
	}

	// Decode client certificate
	certBlock, _ := pem.Decode(clientCertPEM)
	if certBlock == nil {
		return errors.New("failed to decode client certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse client certificate: %w", err)
	}

	// Decode client key
	keyBlock, _ := pem.Decode(clientKeyPEM)
	if keyBlock == nil {
		return errors.New("failed to decode client key PEM")
	}

	clientKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse client key: %w", err)
	}

	// Verify that the public key from the certificate matches the private key
	if err := cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature); err != nil {
		return fmt.Errorf("client certificate signature verification failed: %w", err)
	}

	// Ensure the client certificate and key are compatible
	if _, err := rsa.SignPKCS1v15(nil, clientKey, crypto.Hash(cert.SignatureAlgorithm), cert.Signature); err != nil {
		return errors.New("client certificate and key are not compatible")
	}

	return nil
}

func detectKubeAPIServer(opts AceValidateOptions, restConfig *rest.Config) error {
	kc, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	if host, err := GetAPIServer(kc); err == nil {
		log.Printf("ace cluster kube-apiserver address: %s, source: configmap kube-public/cluster-info\n", host)
	} else {
		log.Printf("failed to get kube-apiserver address from configmap kube-public/cluster-info, reason: %s\n", err)

		sv, err := kc.ServerVersion()
		if err != nil {
			return err
		}

		var done bool
		if strings.Contains(sv.GitVersion, "+k3s") {
			if len(opts.SelfManagement.TargetIPs) > 0 {
				restConfig.Host = fmt.Sprintf("https://%s:6443", opts.SelfManagement.TargetIPs[0])
				done = true
				log.Printf("ace cluster kube-apiserver address: %s, source: selfManagement.TargetIPs[0]\n", restConfig.Host)
			} else {
				serverIP, err := getIngressIP(kc)
				if err != nil {
					return err
				}
				if serverIP != "" {
					restConfig.Host = fmt.Sprintf("https://%s:6443", serverIP)
					done = true
					log.Printf("ace cluster kube-apiserver address: %s, source: service ace/ace-ingress\n", restConfig.Host)
				}
			}
		}
		if !done {
			return fmt.Errorf("failed to detect ace kube-apiserver address\n")
		}
	}
	return nil
}

// GetAPIServer gets the api server url
func GetAPIServer(kubeClient kubernetes.Interface) (string, error) {
	config, err := getClusterInfoKubeConfig(kubeClient)
	if err != nil {
		return "", err
	}
	clusters := config.Clusters
	if len(clusters) != 1 {
		return "", fmt.Errorf("can not find the cluster in the cluster-info")
	}
	cluster := clusters[0].Cluster
	return cluster.Server, nil
}

func getClusterInfoKubeConfig(kubeClient kubernetes.Interface) (*clientcmdapiv1.Config, error) {
	cm, err := kubeClient.CoreV1().ConfigMaps("kube-public").Get(context.TODO(), "cluster-info", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	config := &clientcmdapiv1.Config{}
	err = yaml.Unmarshal([]byte(cm.Data["kubeconfig"]), config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getIngressIP(kc kubernetes.Interface) (string, error) {
	// https://github.com/appscode-cloud/lib-selfhost/commit/6be70cbd48d4083449846d6d2d638247ede08421
	svcLabels := map[string]string{
		"app.kubernetes.io/component": "controller",
		"app.kubernetes.io/instance":  "ace",
		"app.kubernetes.io/name":      "ace-ingress",
	}
	svcList, err := kc.CoreV1().Services(corev1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(svcLabels).String(),
	})
	if err != nil {
		return "", err
	}
	if len(svcList.Items) == 0 {
		return "", errors.New("ace-ingress not found!")
	}

	for _, ing := range svcList.Items[0].Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			return ing.IP, nil
		} else if ing.Hostname != "" {
			return ing.Hostname, nil
		}
	}
	return "", nil
}
