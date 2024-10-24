package installer_precheck

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
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

var allOk = true

func CheckOptions(optsPath string) (bool, error) {
	data, err := os.ReadFile(optsPath)
	if err != nil {
		return false, fmt.Errorf("failed to read options file. Reason: %w", err)
	}

	var aceOptions v1alpha1.AceOptionsSpec
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

	// 1. verify client-cert and client-key
	if aceOptions.Registry.Certs.ClientCert != "" && aceOptions.Registry.Certs.ClientKey != "" {
		if err = checkCertKeyPair(aceOptions.Registry.Certs.ClientCert, aceOptions.Registry.Certs.ClientKey); err != nil {
			allOk = false
			log.Printf("failed to verify client-cert and client-key. Reason: %s", err)
		}
	}

	// 2. check image pull secret exist
	if len(aceOptions.Registry.ImagePullSecrets) > 0 {
		for _, sec := range aceOptions.Registry.ImagePullSecrets {
			_, err := kc.CoreV1().Secrets("kubedb").Get(context.Background(), sec, metav1.GetOptions{})
			if err != nil {
				allOk = false
				if kerr.IsNotFound(err) {
					log.Printf("%s image pull secret not found\n", sec)
				} else {
					log.Printf("failed to get image pull secret: %s. Reason: %s\n", sec, err)
				}
			}
		}

		// 3. check registry credentials
		_, err := authn.ImageWithDigest(kc, fmt.Sprintf("%s/ace-installer", aceOptions.Registry.Image.Proxies.AppsCode), &k8schain.Options{
			ImagePullSecrets: aceOptions.Registry.ImagePullSecrets,
		})
		if err != nil {
			allOk = false
			log.Printf("failed to verify registry credentials. Reason: %s\n", err)
		}
	}

	// 4. check docker registry has https://
	if aceOptions.Registry.Image.Proxies.DockerHub != "" {
		checkDockerRegistry(aceOptions)
	}

	// 5. check kube-apiserver can be detected
	if aceOptions.InitialSetup.SelfManagement.Import {
		if err = detectKubeAPIServer(aceOptions, rc); err != nil {
			allOk = false
			log.Printf("failed to detect kube-apiserver. Reason: %s\n", err)
		}
	}

	// 6. Check cluster_id provided in options matches the actual cluster id for offline installer
	checkLicenses(aceOptions, kc)

	// 7. check disable-features already exists in cluster or not
	checkDisabledFeatures(kc, aceOptions)

	if err = checkDefaultStorageClassExists(kc); err != nil {
		allOk = false
		log.Printf("falied to get default storage class. reason: %s", err)
	}

	return allOk, nil
}

func checkCertKeyPair(certPEM, keyPEM string) error {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil || block.Type != "CERTIFICATE" {
		return errors.New("failed to decode certificate PEM block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	block, _ = pem.Decode([]byte(keyPEM))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return errors.New("failed to decode private key PEM block")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("certificate public key is not RSA")
	}

	if pubKey.N.Cmp(key.N) != 0 || pubKey.E != key.E {
		return errors.New("certificate and private key do not match")
	}

	return nil
}

func detectKubeAPIServer(opts v1alpha1.AceOptionsSpec, restConfig *rest.Config) error {
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
			if len(opts.Infra.DNS.TargetIPs) > 0 {
				restConfig.Host = fmt.Sprintf("https://%s:6443", opts.Infra.DNS.TargetIPs[0])
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

func checkDockerRegistry(aceOptions v1alpha1.AceOptionsSpec) {
	url := "https://" + aceOptions.Registry.Image.Proxies.DockerHub
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		allOk = false
		log.Printf("Error creating request: %v", err)
		return
	}

	resp, clientErr := client.Do(req)
	if clientErr != nil {
		allOk = false
		log.Printf("Error making request: %v", clientErr)
		return
	}
	defer resp.Body.Close()

	if resp.TLS == nil || (resp.TLS != nil && len(resp.TLS.PeerCertificates) == 0) {
		allOk = false
		log.Printf("No TLS connection or no peer certificates.")
	}
}

func checkLicenses(aceOptions v1alpha1.AceOptionsSpec, kc *kubernetes.Clientset) {
	ns, err := kc.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	if err != nil {
		allOk = false
		log.Printf("failed to get kube-system namespace. Reason: %s", err)
		return
	}

	ca, err := info.LoadLicenseCA()
	if err != nil {
		allOk = false
		log.Printf("failed to get license ca. Reason: %s", err)
		return
	}

	caCert, err := info.ParseCertificate(ca)
	if err != nil {
		allOk = false
		log.Printf("failed to parse license ca. Reason: %s", err)
		return
	}

	for pro, lic := range aceOptions.Context.Licenses {
		_, err := verifier.ParseLicense(verifier.ParserOptions{
			ClusterUID: string(ns.UID),
			CACert:     caCert,
			License:    []byte(lic),
		})
		if err != nil {
			allOk = false
			log.Printf("failed to verify license for product: %s. Reason: %s", pro, err)
		}
	}
}

func checkDisabledFeatures(kc *kubernetes.Clientset, aceOptions v1alpha1.AceOptionsSpec) {
	features := map[string]func(*kubernetes.Clientset){
		"kubedb":       checkKubeDBExists,
		"stash":        checkStashExists,
		"kubestash":    checkKubeStashExists,
		"cert-manager": checkCertManagerExists,
	}

	for feature, checkFunc := range features {
		if modstring.Contains(aceOptions.InitialSetup.SelfManagement.DisableFeatures, feature) {
			checkFunc(kc)
		}
	}
}

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

func checkFeatureExists(kc *kubernetes.Clientset, featureName string, labels map[string]string) {
	found, err := checkDeploymentExists(kc, labels)
	if err != nil {
		allOk = false
		log.Printf("failed to get %s deployments. Reason: %s\n", featureName, err)
		return
	}

	if !found {
		allOk = false
		log.Printf("%s not found in this cluster", featureName)
	}
}

func checkKubeDBExists(kc *kubernetes.Clientset) {
	kubedbLabels := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "kubedb-provisioner",
	}
	checkFeatureExists(kc, "kubedb", kubedbLabels)
}

func checkStashExists(kc *kubernetes.Clientset) {
	stashLabels := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "stash-enterprise",
	}
	checkFeatureExists(kc, "stash", stashLabels)
}

func checkKubeStashExists(kc *kubernetes.Clientset) {
	kubeStashLabels := map[string]string{
		"app.kubernetes.io/instance":   "kubestash",
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "kubestash-operator",
	}
	checkFeatureExists(kc, "kubestash", kubeStashLabels)
}

func checkCertManagerExists(kc *kubernetes.Clientset) {
	certManagerLabels := map[string]string{
		"app.kubernetes.io/instance":  "cert-manager",
		"app.kubernetes.io/component": "controller",
		"app.kubernetes.io/name":      "cert-manager",
	}

	// Check for cert-manager deployment
	checkFeatureExists(kc, "cert-manager", certManagerLabels)

	// Check for cert-manager webhook deployment
	webhookLabels := map[string]string{
		"app.kubernetes.io/instance":  "cert-manager",
		"app.kubernetes.io/component": "webhook",
		"app.kubernetes.io/name":      "webhook",
	}
	checkFeatureExists(kc, "cert-manager webhook", webhookLabels)
}

func checkDefaultStorageClassExists(kc *kubernetes.Clientset) error {
	storageClasses, err := kc.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, sc := range storageClasses.Items {
		if val, exists := sc.Annotations["storageclass.kubernetes.io/is-default-class"]; exists && val == "true" {
			return nil
		}
	}

	return errors.New("default storage-class not found")
}
