package cli

import (
	"fmt"
	"bytes"
	"kubeokta/kubeconfig"
	"github.com/jessevdk/go-flags"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/tools/clientcmd/api"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"log"
)

type kconfig *api.Config

// TokenSet structure for getting okta response
type TokenSet struct {
	IDToken			string `json:"id_token"`
	RefreshToken	string `json:"refresh_token"`
	AccessToken		string `json:"access_token"`
}

// CliParameters structure for getting right flags
type CliParameters struct {
	KubernetesCluster	string	`long:"cluster" env:"K8S_CLUSTER" description:"Kubernetes cluster for okta authentication."`
	OktaUser			string	`long:"okta-user" env:"OKTA_USER" description:"Okta user for authentication."`
	IssuerURL			string	`long:"issuer-url" env:"ISSUER_URL" description:"Issuer URL of your okta authorization server."`
	ClientID			string	`long:"client-id" env:"CLIENT_ID" description:"Client ID of your OIDC Okta application."`
	ClientSecret		string	`long:"client-secret" env:"CLIENT_SECRET" description:"CLient Secret of your OIDC Okta application."`
}

//Parse arguments
func Parse(args []string) (*CliParameters, error) {
	var cli CliParameters

	parser := flags.NewParser(&cli, flags.HelpFlag)
	parser.LongDescription = fmt.Sprint("This package is for okta authentication inside kubernetes. For more information you can check https://github.com/campanate/kubeokta .")
	args, err := parser.ParseArgs(args[1:])

	if err != nil {
		return nil, err
	}

	return &cli, nil

}

// Execute the script
func Execute(cli CliParameters) error {
	path, err := homedir.Expand("~/.kube/config")

	if err != nil {
		return err
	}

	kconfig, err := clientcmd.LoadFromFile(path)

	if err != nil {
		return err
	}


	if cli.OktaUser == "" || cli.IssuerURL == "" || cli.ClientID == "" || cli.ClientSecret == "" || cli.KubernetesCluster == "" {
		return fmt.Errorf(`please pass the flags --okta-user, --cluster, --issuer-url, --client-id and --client-secret`)
	}

	kubeconfig.CreateContext(cli.OktaUser, cli.KubernetesCluster, kconfig)
	kubeconfig.CreateOktaConfig(cli.OktaUser, cli.IssuerURL, cli.ClientID, cli.ClientSecret, kconfig)

	err = clientcmd.WriteToFile(*kconfig, path)

	if err != nil {
		return err
	}

	fmt.Printf("Now, please type your okta password:\n")
	password, err := terminal.ReadPassword(0)

	if string(password) == "" {
		return fmt.Errorf("Error: Password can not be empty")
	}

	resp, err := GetResponseToken(cli, string(password))

	if err != nil {
		return fmt.Errorf("Error: %s", err.Error())
	}

	err = HandleResponseToken(resp, kconfig)

	if err != nil {
		return fmt.Errorf("Error: %s", err.Error())
	}

	return nil

}


// GetResponseToken from Okta API
func GetResponseToken(cli CliParameters, password string) (*http.Response, error) {

	client := &http.Client{
		Timeout: time.Second * time.Duration(5),
	}

	metadata, err := getMetaData(cli.IssuerURL)

	if err != nil {
		return nil, err
	}

	reqBody := []byte("client_id=" + cli.ClientID +
			"&client_secret=" + cli.ClientSecret +
			"&username=" + cli.OktaUser +
			"&grant_type=password" +
			"&password=" + password +
			"&scope=openid profile")

	req, err := http.NewRequest("POST", metadata["token_endpoint"].(string), bytes.NewBuffer(reqBody))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
} 

// HandleResponseToken from Okta API
func HandleResponseToken(resp *http.Response, config *api.Config) error {
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		
		token := &TokenSet{}

		err = json.Unmarshal(bodyBytes, &token)
		if err != nil {
			return fmt.Errorf("Error: %s", err.Error())
		}

		authProvider := kubeconfig.GetAuthProvider(config)
		
		authProvider.Config["id-token"] = token.IDToken
		authProvider.Config["refresh-token"] = token.RefreshToken
		authProvider.Config["access-token"] = token.AccessToken
		
		path, err := homedir.Expand("~/.kube/config")

		if err != nil {
			return err
		}


		err = clientcmd.WriteToFile(*config, path)

		log.Printf("Updated %s", path)

	} else {
		errorMsg := make(map[string]interface{})

		err = json.Unmarshal(bodyBytes, &errorMsg)

		if err != nil {
			return fmt.Errorf("Error: %s", err.Error())
		}

		return fmt.Errorf("%s", errorMsg["error_description"].(string))
	} 

	return nil
}

//getMetaData gets the token endpoint
func getMetaData(url string) (map[string]interface{}, error) {
	metaDataURL := url + "/.well-known/openid-configuration"

	resp, err := http.Get(metaDataURL)

	if err != nil {
		return nil, fmt.Errorf("request for metadata was not successful: %s", err.Error())
	}

	defer resp.Body.Close()

	md := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&md)

	return md, nil
}