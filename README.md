# AWSets

A utility for crawling an AWS account and exporting all its resources for further analysis.

## Badges
[![Release](https://img.shields.io/github/v/release/trek10inc/awsets?include_prereleases&style=for-the-badge)](https://github.com/trek10inc/awsets/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Build status](https://img.shields.io/github/workflow/status/trek10inc/awsets/test?style=for-the-badge)](https://github.com/trek10inc/awsets/actions?workflow=test)

## Motivation
Trek10 frequently gets pulled into existing AWS accounts that lack documentation, don’t practice proper tagging, don’t use infrastructure as code, or just contain so many resources that it is difficult to get an understanding of what we’re working with. Unfortunately, there is no single AWS call or service that can provide a complete assessment of everything in an account so that we can start to piece together a map of what is going on.

After exploring existing solutions in this space, we were unable to find anything that both had the resource coverage we desired, and also aligned with the goals we set out with.

## Goals
This project has two main goals:
* Support as many AWS resources as possible
* Build relationships between those resources
* Normalize output to facilitate post-processing.

## Notes
* global resources (`iam`, `route53`, `waf`) are always queried regardless of region filter as long as the resource type is valid
* Not every resource has support yet, not every resource has tags yet, and not all relationships are in place. If a gap in functionality has been identified, please submit a request to have it fixed/added.

The output of this tool is a JSON array of objects in the following format:
```json5
{
    "Account": "123456789",              // account resource is in
    "Region": "us-east-1",               // region resource is in
    "Id": "12345",                       // resource id
    "Version": "",                       // resource version
    "Type": "ec2/instance",              // resource type
    "Name": "test-instance",             // resource name
    "Attributes": {},                    // full dump of resource attributes
    "Tags": {},                          // normalized tags for resource
    "Relations": [                       // array of the identifiers of related resources
        {
        "Account": "123456789",
        "Region": "us-east-1",
        "Id": "vpc-123abc123",
        "Version": "",
        "Type": "ec2/vpc"
        }
    ]
}
```

Filters can be added to the query in order to restrict regions and resource types. A list of currently supported AWS resource types can be found [here](supported_resources.txt).

## Getting Started

### Installation
#### From source
```
git clone https://github.com/trek10inc/awsets.git
cd awsets/cmd/awsets
go build && go install
```

#### Homebrew
```
brew tap trek10inc/tap
brew install awsets
```

#### From binaries
Binaries are available [here](https://github.com/trek10inc/awsets/releases)

## Usage:
```
USAGE:
   awsets [global options] command [command options] [arguments...]

COMMANDS:
   list     lists all requested aws resources
   regions  lists regions supported by account
   types    lists supported resource types
   process  runs processors on results json
   version  prints version information
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

Region Filter:
This is a comma separated list of region prefixes. For example, `--regions us-e` would run in `us-east-1` and `us-east-2`. `--regions us-e,us-west-1` would run in `us-east-1`, `us-east-2`, and `us-west-1`. `--regions all` will run in all regions.

Resource filter:
This is broken into two flags: `--include` and `--exclude`. Both are comma-delimited list of resource types, with the exclusions processing last.

`awsets list --include iam` will query `iam/group`, `iam/instanceprofile`, `iam/policy`, `iam/role`, and `iam/user`

`awsets list --include iam --exclude iam/g` will query `iam/instanceprofile`, `iam/policy`, `iam/role`, and `iam/user`

### Subcommands

#### list

Primary command - used to do the actual query.

```
USAGE:
   awsets list [command options]

OPTIONS:
   --dryrun                  do a dry run of query (default: false)
   --include value           comma separated list of resource type prefixes to include
   --exclude value           comma separated list of resource type prefixes to exclude
   --output value, -o value  output file to save results
   --profile value           AWS profile to use
   --refresh                 force a refresh of cache (default: false)
   --regions value           comma separated list of region prefixes
   --show-progress           toggle progress bar (default: false)
   --verbose, -v             toggle verbose logging (default: false)
   --help, -h                show help (default: false)
```

Examples:

Query everything, save to `all.json`
`awsets list -o all.json`

Query all resources managed by the IAM & EC2 services in us-east-1:
`awsets list --regions us-east-1 -o all.json --include iam,ec2`

#### regions

Simple command to output all supported and enabled regions for the current AWS account. The arguments are used to filter the regions by prefix.

```
USAGE:
   awsets regions [command options] [region prefixes]

OPTIONS:
   --profile value  AWS profile to use
   --help, -h  show help (default: false)
```

#### types

Simple command to output all supported AWS resource types. Flags can be passed in include/exclude specific resource types by prefix.

```
USAGE:
   awsets types [command options]  

OPTIONS:
   --include value  comma separated list of resource type prefixes to include
   --exclude value  comma separated list of resource type prefixes to exclude
   --help, -h       show help (default: false)
```

#### process

A section of experimentation. There are a few custom processors here that are used to manipulate the output `awsets` json. Most will likely be split out to be separate applications or scripts, but for ease of development have been placed here. Long term, this may still contain general utilities to help search and organize the data (like a DOT graph builder?, stats), but will not contain specialized analysis (Cloudformation healthcheck) or anything that can already be done better by other CLI tools like `jq`. 

##### dot

Command that takes a file that is output from `awsets list` and generates a DOT graph. This can then be rendered into an image via `fdp <dot file> -Tsvg -o <output.svg>`. This step can take a while to complete, and the resulting image is typically rather large. There is ongoing work to try and improve this process.

```
USAGE:
   awsets process dot [command options] [arguments...]

OPTIONS:
   --input value, -i value   input file containing data to process
   --output value, -o value  output file to save results
   --hide-unrelated          remove unrelated resources from dot file (default: false)
   --help, -h                show help (default: false)
```

## Future Work

Although AWSets is in a place where it provides solid resource coverage and works well for a lot of use cases, there is more work to be done:
* Supporting more AWS resources and relationships - 300+ is a good start, but there are many more to go
* In addition to supporting more resources, existing resources may have some gaps. For example, some resources require secondary calls to get Tags
* Improve relationship building - AWSets should be able to match a DynamoDB table to a Lambda Function when the DDB table is passed in via environment variable
