package kubeconfig

import (
	"k8s.io/client-go/tools/clientcmd/api"
)

// GetAuthProvider returns the current provider context
func GetAuthProvider(config *api.Config) (*api.AuthProviderConfig) {
	context := config.Contexts[config.CurrentContext]
	authInfo := config.AuthInfos[context.AuthInfo]

	return (*api.AuthProviderConfig)(authInfo.AuthProvider)
	
}

// CreateContext creates a new context based on parameters.
func CreateContext(user string, cluster string, config *api.Config) {
	newCluster := api.NewCluster()
	newCluster.Server = cluster

	clusters := map[string]*api.Cluster{}
	clusters[cluster] = newCluster

	config.Clusters = clusters

	context := api.NewContext()
	context.Namespace = user
	context.AuthInfo = user
	context.Cluster = cluster

	contexts := map[string]*api.Context{
		user: context,
	}

	config.Contexts = contexts

	config.CurrentContext = user

}

// CreateOktaConfig creates a OIDC configuration in your kube config file
func CreateOktaConfig(user string, issuerURL string, clientID string, clientSecret string, cfg *api.Config) {
	auth := api.NewAuthInfo()

	oktaConfig := map[string]string{
		"idp-issuer-url": issuerURL,
		"client-id":      clientID,
		"client-secret":  clientSecret,
	}

	authProvider := &api.AuthProviderConfig{
		Name:   "oidc",
		Config: oktaConfig,
	}

	auth.AuthProvider = authProvider

	v := map[string]*api.AuthInfo{
		user: auth,
	}

	cfg.AuthInfos = v

}