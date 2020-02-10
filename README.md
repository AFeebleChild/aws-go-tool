# aws-go-tool

## Flags

### Access Type Flag

Required

The "-a" flag needs to be one of the following:
- role
- profile
- instance

If the flag is `role`, then the tool assumes that a list of cross account role names are going to be passed in.  They need to be configured in the shared config file `~/.aws/config` as follows:
```
[profile <profileName>]
role_arn = arn:aws:iam::123456789012:role/<roleName>
source_profile = saml
region = us-east-1
output = json
```

The `source_profile = saml` is also required, as that profile name is hardcoded into the tool as of now.  This is a profile that needs to be configured in your `~/.aws/credentials` file, and have access to assume the list of roles passed into the tool.

If the flag is `profile`, then the list should be profiles setup in your `~/.aws/credentials` file.

If the flag is `instance`, then the script needs to be run from an instance with cross account access to the config file you pass in.

### Profiles Flag

Required

The "-p" flag needs to be passed as a text file, with 1 profile per line.

```
profile1
profile2
profile3
profile4
```

### Tags Flag

Optional

The "-g" flag can be passed as a text file with one tag name per line, to be added to the csv output.

```
Name1
Name2
Name3
Name4
```

## Supported Commands
- EC2
    - `instanceslist`
    - `volumeslist`
    - `snapshotslist`
    - `imagelist`
    - `imagecheck`
        - Checks the images in the account for any the are in use by the instances, and how many use it.  It does not check for the AMI being shared to other accounts.
    - `sgslist`
    - `sgruleslist`
- IAM
    - `policieslist`
    - `roleslist`
    - `rolesupdate`
        - Still in progress, but updates the list of roles to an 8 hour assume role duration
    - `userslist`
    - `userupdatepw`
        - Use the "-u" flag to pass in the username you wish to update the password for.
- S3
    - `bucketslist`
    - `filesize`
        - This is used to gather the size of all file types in the specified buckets. Either a file with a list of buckets can be entered with the `-b` flag, or "public-only" can be passed to the same flag. If a file with a list of buckets is passed, it will only search the first account in the profiles file passed. If the public-only parameter is passed, it will search through all the accounts in the profiles file.
- VPC
    - `vpcslist`
    - `subnetslist`
- Workspaces (in progress)

### TODO
- Rework utils.BuildAccountsSlice to be a global variable, and not needed to be called on every cli command.
- Add printer function for csv
- Add global option for json or csv output
- Add utils file in cmd to have an easier wrapper func to build account list, get tags, logging, etc.

### Download Links
Linux - https://afeeblechild.s3-us-west-2.amazonaws.com/go-binaries/aws-go-tool-linux.zip

Windows - https://afeeblechild.s3-us-west-2.amazonaws.com/go-binaries/aws-go-tool-windows.zip

Mac - https://afeeblechild.s3-us-west-2.amazonaws.com/go-binaries/aws-go-tool-mac.zip
