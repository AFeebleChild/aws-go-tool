# aws-go-tool

### Access Type flag
The "-a" flag needs to be either `role` or `profile`.  If it is `role`, then the tool assumes that a list of cross account role names
are going to be passed in.  They need to be configured in the shared config file `~/.aws/config` as follows:
```
[profile <profileName>]
role_arn = arn:aws:iam::123456789012:role/<roleName>
source_profile = saml
region = us-east-1
output = json
```

The `source_profile = saml` is also required, as that profile name is hardcoded into the tool as of now.  This
is a profile that needs to be configured in your `~/.aws/credentials` file, and have access to assume the list of roles
passed into the tool.




# TODO
Add printer function for csv
