package client

import (
	"flag"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
	"path/filepath"
)

type LocalDevMode bool

type K8sClient struct {
	devMode LocalDevMode
	config  *rest.Config
}

func NewK8sClient(devMode LocalDevMode) (*K8sClient, error) {
	config, err := getKubeConfig(devMode)
	if err != nil {
		return nil, err
	}
	return &K8sClient{
		devMode: devMode,
		config:  config,
	}, nil
}

//TODO use it as generic function across system
func getKubeConfig(devMode LocalDevMode) (*rest.Config, error) {
	if devMode {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		kubeconfig := flag.String("kubeconfig-authenticator", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}
		return restConfig, nil
	} else {
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return restConfig, nil
	}
}

func (impl *K8sClient) GetArgoConfig() (secret *v1.Secret, cm *v1.ConfigMap, err error) {
	clientSet, err := kubernetes.NewForConfig(impl.config)
	if err != nil {
		return nil, nil, err
	}
	secret, err = clientSet.CoreV1().Secrets(ArgocdNamespaceName).Get(ArgoCDSecretName, v12.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	cm, err = clientSet.CoreV1().ConfigMaps(ArgocdNamespaceName).Get(ArgoCDConfigMapName, v12.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	return secret, cm, nil
}

// argocd specific conf
const (
	SettingAdminPasswordHashKey = "admin.password"
	// SettingAdminPasswordMtimeKey designates the key for a root password mtime inside a Kubernetes secret.
	SettingAdminPasswordMtimeKey = "admin.passwordMtime"
	SettingAdminEnabledKey       = "admin.enabled"
	SettingAdminTokensKey        = "admin.tokens"

	SettingServerSignatureKey = "server.secretkey"
	settingURLKey             = "url"
	ArgoCDConfigMapName       = "argocd-cm"
	ArgoCDSecretName          = "argocd-secret"
	ArgocdNamespaceName       = "devtroncd"
)

func (impl *K8sClient) GetServerSettings() (*DexConfig, error) {
	cfg := &DexConfig{}
	secret, cm, err := impl.GetArgoConfig()
	if err != nil {
		return nil, err
	}
	if settingServerSignatur, ok := secret.Data[SettingServerSignatureKey]; ok {
		cfg.ServerSecret = string(settingServerSignatur)
	}
	if settingURL, ok := cm.Data[settingURLKey]; ok {
		cfg.Url = settingURL
	}
	return cfg, nil
}
