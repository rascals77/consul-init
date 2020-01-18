package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"text/template"

	consulAPI "github.com/hashicorp/consul/api"
	"github.com/rascals77/consul-init/config"
	"github.com/rascals77/consul-init/consul"
	"github.com/rascals77/consul-init/util"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
)

var (
	configFile = flag.String("config", "", "Config file")
	bootstrap  = flag.Bool("bootstrap", false, "Perform ACL bootstrap")
)

func main() {
	flag.Parse()

	if *configFile == "" {
		log.Fatal("the -config parameter is required")
	}

	// Verify config file exists
	if util.IsNotExist(*configFile) {
		log.Fatalf("config file [%s] does not exist", *configFile)
	}

	log.Printf("using config file [%s]", *configFile)

	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error reading config file [%s]", err)
	}

	var conf config.Config
	err := viper.Unmarshal(&conf)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	validate := validator.New()
	err = validate.Struct(conf)
	if err != nil {
		log.Fatal(err)
	}

	// verify token secrets file does not already exist
	if util.IsExist(conf.TokenSecretsFile) {
		log.Fatalf("file named [%s] already exists", conf.TokenSecretsFile)
	}

	consulURL := fmt.Sprintf("%s:%d", conf.Address, conf.Port)
	consulConfig := consulAPI.DefaultConfig()
	consulConfig.Address = consulURL
	consulConfig.Scheme = conf.Scheme
	//consulConfig.Token = bootstrapToken
	consulConfig.TLSConfig = consulAPI.TLSConfig{
		CAFile:             conf.CACert,
		CertFile:           conf.Cert,
		KeyFile:            conf.Key,
		InsecureSkipVerify: false,
	}

	// create Consul client
	var consulClient *consulAPI.Client
	/*
		log.Printf("connecting to Consul running at [%s://%s]", conf.ConsulScheme, consulURL)
		consulClient, err = consulAPI.NewClient(consulConfig)
		if err != nil {
			log.Fatal(err)
		}
	*/

	// handle bootstrapping and obtaining bootstrap token
	var bootstrapToken string
	if *bootstrap == true {
		// verify bootstrap token file does not already exist
		if util.IsExist(conf.TokenFile) {
			log.Fatalf("file named [%s] already exists", conf.TokenFile)
		}

		// bootstrap Consul
		log.Printf("performing ACL bootstrapping")
		log.Printf("connecting to Consul running at [%s://%s]", conf.Scheme, consulURL)
		consulClient, err = consulAPI.NewClient(consulConfig)
		if err != nil {
			log.Fatal(err)
		}
		bootstrapToken, err = consul.BootstrapACL(consulClient)
		if err != nil {
			log.Fatal(err)
		}

		if bootstrapToken == "" {
			log.Fatalf("unable to obtain bootstrap token")
		}

		// write bootstrap token to file
		bootstrapTokenWithNewline := append([]byte(bootstrapToken), "\n"...)
		err = ioutil.WriteFile(conf.TokenFile, []byte(bootstrapTokenWithNewline), 0600)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// get bootstrap token from file
		log.Printf("getting bootstrap token from [%s] file", conf.TokenFile)
		bootstrapToken, err = util.GetFirstLineOfFile(conf.TokenFile)
		if err != nil {
			log.Fatal(err)
		}

		if bootstrapToken == "" {
			log.Fatalf("unable to obtain bootstrap token")
		}
	}

	// add the token to the Consul API client configuration
	consulConfig.Token = bootstrapToken
	connectingMsg := fmt.Sprintf("connecting to Consul running at [%s://%s]", conf.Scheme, consulURL)
	if *bootstrap == true {
		log.Printf("re-%s", connectingMsg)
	} else {
		log.Printf(connectingMsg)
	}
	consulClient, err = consulAPI.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}

	// get current list of policies
	log.Printf("getting currect list of ACL policies")
	policies, err := consul.GetPolicies(consulClient)
	if err != nil {
		log.Fatal(err)
	}

	// generate node agent policies and add them to policies that need to exist
	for _, node := range conf.Members {
		memberAgentPolicy := config.Policies{
			Name:        fmt.Sprintf("%s-agent", node.Name),
			Description: fmt.Sprintf("%s agent token", node.Name),
		}
		var templateOut bytes.Buffer
		t := template.Must(template.New("nodeAgentRule").Parse(conf.NodeAgentTemplate))
		err = t.Execute(&templateOut, node)
		if err != nil {
			log.Fatal(err)
		}
		memberAgentPolicy.Rules = templateOut.String()
		conf.Policies = append(conf.Policies, memberAgentPolicy)
	}

	// create policies, if needed
	for _, policy := range conf.Policies {
		policyName := strings.TrimSpace(policy.Name)
		policyAlreadyExists := consul.IsPolicyNameExist(policies, policyName)
		if policyAlreadyExists == false {
			log.Printf("creating policy named [%s]", policyName)
			newPolicy := consul.NewPolicy{
				Name:        policyName,
				Description: strings.TrimSpace(policy.Description),
				Rules:       strings.TrimSpace(policy.Rules),
			}
			err = consul.CreatePolicy(consulClient, newPolicy)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Printf("policy named [%s] already exists", policyName)
		}
	}

	// get current list of tokens
	log.Printf("getting currect list of ACL tokens")
	tokens, err := consul.GetTokens(consulClient)
	if err != nil {
		log.Fatal(err)
	}

	// generate node agent tokens and add them to tokens that need to exist
	for _, node := range conf.Members {
		n := fmt.Sprintf("%s-agent", node.Name)
		nodeToken := config.Tokens{
			Name:     n,
			Policies: []string{n},
		}
		conf.Tokens = append(conf.Tokens, nodeToken)
	}

	// create tokens, if needed
	//tokenSecretIDs := map[string]string{}
	tokenSecretIDs := make(map[string]string)
	for _, token := range conf.Tokens {
		var secretID string
		tokenName := strings.TrimSpace(token.Name)
		tokenAlreadyExists, tokenSecretID := consul.IsTokenDescExist(tokens, tokenName)
		if tokenAlreadyExists == false {
			log.Printf("creating token with description [%s]", tokenName)
			newToken := consul.NewToken{
				Description: tokenName,
				PolicyNames: token.Policies,
			}
			secretID, err = consul.CreateToken(consulClient, newToken)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Printf("token with description [%s] already exists", tokenName)
			secretID = tokenSecretID
		}
		tokenSecretIDs[token.Name] = secretID
	}

	// create JSON of token secrets
	jsonTokenSecretIDs, err := json.MarshalIndent(tokenSecretIDs, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// write JSON to disk
	log.Printf("writing token secrets to [%s]", conf.TokenSecretsFile)
	jsonTokenSecretIDsWithNewline := append([]byte(jsonTokenSecretIDs), "\n"...)
	err = ioutil.WriteFile(conf.TokenSecretsFile, []byte(jsonTokenSecretIDsWithNewline), 0600)
	if err != nil {
		log.Fatal(err)
	}

	// set agent ACL tokens for each member
	for _, node := range conf.Members {
		// set client configuration for consul member
		consulURL := fmt.Sprintf("%s:%d", node.IP, node.Port)
		consulConfig.Address = consulURL
		log.Printf("connecting to Consul running at [%s://%s]", conf.Scheme, consulURL)
		consulClient, err = consulAPI.NewClient(consulConfig)
		if err != nil {
			log.Fatal(err)
		}

		// get the agent token to set on consul member
		n := fmt.Sprintf("%s-agent", node.Name)
		memberToken, found := tokenSecretIDs[n]
		if found == false {
			log.Fatalf("agent token [%s] was not found", n)
		}

		// set the agent token on the consul member
		log.Printf("setting the ACL agent token on [%s]", node.Name)
		err = consul.SetAgentACLToken(consulClient, memberToken)
		if err != nil {
			log.Fatal(err)
		}
	}
}
