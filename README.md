## Compile

To compile ```consul-init``` run the following command:

```
$ mkdir -p ~/go/src/github.com/rascals77
$ cd ~/go/src/github.com/rascals77
$ git clone https://github.com/rascals77/consul-init.git
$ cd consul-init
$ go mod download
$ CGO_ENABLED=0 go build -ldflags="-s -w" -o consul-init
```

## Execute

##### Prerequisites

- Consul has been installed and running on all members of the cluster
- To use the ```-bootstrap``` option, the consul member ```consul-init``` is executed on has to be running with ```bootstrap: true```

To execute the application using command line flags:

```
$ ./consul-init -h
Usage of ./consul-init:
  -bootstrap
    	Perform ACL bootstrap
  -config string
    	Config file
```

To bootstrap a newly created Consul cluster, run the following command.  The master token and ACL token SecretIDs are written to the files provided within the ```config.yml``` file.

```
$ ./consul-init -config config.yml -bootstrap
2020/01/18 15:08:51 using config file [config.yml]
2020/01/18 15:08:51 performing ACL bootstrapping
2020/01/18 15:08:51 connecting to Consul running at [https://127.0.0.1:8500]
2020/01/18 15:08:51 re-connecting to Consul running at [https://127.0.0.1:8500]
2020/01/18 15:08:51 getting currect list of ACL policies
2020/01/18 15:08:51 creating policy named [operator-ui]
2020/01/18 15:08:51 creating policy named [vault-agent]
2020/01/18 15:08:51 creating policy named [consul-server-1-agent]
2020/01/18 15:08:51 creating policy named [consul-server-2-agent]
2020/01/18 15:08:51 creating policy named [consul-server-3-agent]
2020/01/18 15:08:51 getting currect list of ACL tokens
2020/01/18 15:08:51 creating token with description [operator-ui]
2020/01/18 15:08:51 creating token with description [vault-agent]
2020/01/18 15:08:51 creating token with description [consul-server-1-agent]
2020/01/18 15:08:51 creating token with description [consul-server-2-agent]
2020/01/18 15:08:51 creating token with description [consul-server-3-agent]
2020/01/18 15:08:51 writing token secrets to [/root/token-secrets.json]
2020/01/18 15:08:51 connecting to Consul running at [https://10.10.10.11:8500]
2020/01/18 15:08:51 setting the ACL agent token on [consul-server-1]
/01/18 15:08:51 connecting to Consul running at [https://10.10.10.12:8500]
2020/01/18 15:08:51 setting the ACL agent token on [consul-server-2]
2020/01/18 15:08:51 connecting to Consul running at [https://10.10.10.13:8500]
2020/01/18 15:08:51 setting the ACL agent token on [consul-server-3]
```

When the Consul cluster has already been bootstrapped and the ```token_secrets_file``` file does not exist:

```
$ ./consul-init -config config.yml
2020/01/18 15:13:26 using config file [config.yml]
2020/01/18 15:13:26 getting bootstrap token from [/root/master-token] file
2020/01/18 15:13:26 connecting to Consul running at [https://127.0.0.1:8500]
2020/01/18 15:13:26 getting currect list of ACL policies
2020/01/18 15:13:26 policy named [operator-ui] already exists
2020/01/18 15:13:26 policy named [vault-agent] already exists
2020/01/18 15:13:26 policy named [consul-server-1-agent] already exists
2020/01/18 15:13:26 policy named [consul-server-2-agent] already exists
2020/01/18 15:13:26 policy named [consul-server-3-agent] already exists
2020/01/18 15:13:26 getting currect list of ACL tokens
2020/01/18 15:13:26 token with description [operator-ui] already exists
2020/01/18 15:13:26 token with description [vault-agent] already exists
2020/01/18 15:13:26 token with description [consul-server-1-agent] already exists
2020/01/18 15:13:26 token with description [consul-server-2-agent] already exists
2020/01/18 15:13:26 token with description [consul-server-3-agent] already exists
2020/01/18 15:13:26 writing token secrets to [/root/token-secrets.json]
2020/01/18 15:13:26 connecting to Consul running at [https://10.10.10.11:8500]
2020/01/18 15:13:26 setting the ACL agent token on [consul-server-1]
2020/01/18 15:13:26 connecting to Consul running at [https://10.10.10.12:8500]
2020/01/18 15:13:26 setting the ACL agent token on [consul-server-2]
2020/01/18 15:13:26 connecting to Consul running at [https://10.10.10.13:8500]
2020/01/18 15:13:26 setting the ACL agent token on [consul-server-3]
```

ACL policies are created for each individual Consul member and then an ACL token is created for each Consul member.

An "operator" token is generated with the ```operator-ui``` policy linked to it.  This is useful for when consul is running with a default policy of ```deny``` (```default_policy = "deny"```), in which case the anonymous token will not have access to anything in the UI.

A token is created for Vault to use this Consul cluster for its backend storage.
