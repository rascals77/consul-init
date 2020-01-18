package consul

// Member defines a Consul member's server name, IPv4 address and TCP port
type Member struct {
	Name string
	IP   string
	Port int
}

// NewPolicy defines a new ACL policy that will be created
type NewPolicy struct {
	Name        string
	Description string
	Rules       string
}

// NewToken defines a new ACL token that will be created.  A token does not have a name, rather just a description.
type NewToken struct {
	Description string
	PolicyNames []string
}

// Policy defines an already existing ACL policy
type Policy struct {
	ID          string
	Name        string
	Description string
}

// TokenDescs defines the details of an already existing ACL token
type TokenDescs struct {
	Desc       string
	AccessorID string
	SecretID   string
}
