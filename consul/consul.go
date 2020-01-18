package consul

import (
	"fmt"

	consulAPI "github.com/hashicorp/consul/api"
)

// IsTokenDescExist returns true and the SecretID string if the given
// token description is in the given TokenDescs array
func IsTokenDescExist(tokens []TokenDescs, val string) (bool, string) {
	for _, t := range tokens {
		if t.Desc == val {
			return true, t.SecretID
		}
	}

	return false, ""
}

// IsPolicyNameExist returns true if the given policy name is
// contained in the given list of policies
func IsPolicyNameExist(policies []Policy, val string) bool {
	for _, p := range policies {
		if p.Name == val {
			return true
		}
	}

	return false
}

// BootstrapACL initializes the ACL system and returns the master token
func BootstrapACL(c *consulAPI.Client) (string, error) {
	aclClient := c.ACL()
	token, _, err := aclClient.Bootstrap()
	if err != nil {
		return "", err
	}

	return token.SecretID, err
}

// GetTokenSecretID returns the SecretID of the token with the given AccessorID
func GetTokenSecretID(c *consulAPI.Client, accessorID string) (string, error) {
	aclClient := c.ACL()
	tokenDetails, _, err := aclClient.TokenRead(accessorID, nil)
	if err != nil {
		return "", err
	}

	return tokenDetails.SecretID, err
}

// GetTokens returns the list of ACL tokens
func GetTokens(c *consulAPI.Client) ([]TokenDescs, error) {
	var tokenDescs []TokenDescs

	aclClient := c.ACL()
	tokenList, _, err := aclClient.TokenList(nil)
	for _, t := range tokenList {
		description := t.Description
		accessorID := t.AccessorID
		secretID, err := GetTokenSecretID(c, accessorID)
		if err != nil {
			return []TokenDescs{}, fmt.Errorf("unable to get SecretID of token with descriptiong [%s]", description)
		}
		tokenDesc := TokenDescs{
			Desc:       description,
			AccessorID: accessorID,
			SecretID:   secretID,
		}
		tokenDescs = append(tokenDescs, tokenDesc)
	}

	return tokenDescs, err
}

// GetPolicies returns the list of ACL policies
func GetPolicies(c *consulAPI.Client) ([]Policy, error) {
	var policies []Policy

	aclClient := c.ACL()
	policyList, _, err := aclClient.PolicyList(nil)
	for _, p := range policyList {
		policy := Policy{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
		policies = append(policies, policy)
	}

	return policies, err
}

// CreatePolicy creates an ACL policy using the parameters given in NewPolicy
func CreatePolicy(c *consulAPI.Client, policy NewPolicy) error {
	aclClient := c.ACL()
	p := consulAPI.ACLPolicy{
		Name:        policy.Name,
		Description: policy.Description,
		Rules:       policy.Rules,
	}
	_, _, err := aclClient.PolicyCreate(&p, nil)

	return err
}

// CreateToken creates a token using the parameters given in NewToken
func CreateToken(c *consulAPI.Client, token NewToken) (string, error) {
	// construct policies
	var policies []*consulAPI.ACLTokenPolicyLink
	for _, policyName := range token.PolicyNames {
		//p := &consulAPI.ACLLink{
		p := &consulAPI.ACLTokenPolicyLink{
			Name: policyName,
		}
		policies = append(policies, p)
	}
	// construct the configuration of the new token
	newTokenConf := consulAPI.ACLToken{
		Description: token.Description,
		Policies:    policies,
	}

	// create the new token
	aclClient := c.ACL()
	newToken, _, err := aclClient.TokenCreate(&newTokenConf, nil)
	if err != nil {
		return "", err
	}

	return newToken.SecretID, err
}

// SetAgentACLToken sets the agent token on the given Client (consul member)
func SetAgentACLToken(c *consulAPI.Client, token string) error {
	agentClient := c.Agent()
	_, err := agentClient.UpdateAgentACLToken(token, nil)

	return err
}
