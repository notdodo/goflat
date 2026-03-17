# goflat

[![Golang CI](https://github.com/notdodo/goflat/actions/workflows/go-ci.yml/badge.svg)](https://github.com/notdodo/goflat/actions/workflows/go-ci.yml)

Flatten complex JSON structures and Go structs (including pointer-heavy types like AWS SDK responses) into a one-dimensional `map[string]interface{}` with dot-notation keys.

`goflat` supports the flattening of:

- JSON strings (via `FlatJSON` or `FlatJSONToMap`)
- Go structs with nested pointers, slices, and maps (via `FlatStruct`)
- JSON strings embedded inside struct fields (e.g. AWS inline policies stored as `*string`)

## Configuration

```golang
goflat.FlattenerConfig{
    Prefix:      "",    // Prepend to all keys
    Separator:   ".",   // Delimiter between path segments (default: ".")
    OmitEmpty:   true,  // Skip empty/zero values
    OmitNil:     true,  // Skip nil pointers and null JSON values
    KeysToLower: false, // Convert all keys to lowercase
}
```

## Examples

### Using a basic JSON structure

```golang
package main

import (
    "fmt"
    "log"

    goflat "github.com/notdodo/goflat/v2"
)

func main() {
    flattened, err := goflat.FlatJSON(`{"a": "3", "b": {"c":true, "a": "", "e": null}}`, goflat.FlattenerConfig{
        Separator: ".",
        OmitEmpty: false,
        OmitNil:   false,
    })
    if err != nil {
        log.Fatalln(err)
    }

    fmt.Println(flattened)
}
```

Output: `{"a":"3","b.a":"","b.c":true,"b.e":null}`

To also remove the `null` values:

```golang
goflat.FlattenerConfig{
    Separator: ".",
    OmitEmpty: false,
    OmitNil:   true,
}
```

### Arrays

When dealing with arrays and recursive structures the library handles depth using indexes:

Input: `[{"a": "3"}, {"b": "3", "C": [{"c": 10}, {"d": 11}]}]`

Output: `{"0.a":"3","1.C.0.c":10,"1.C.1.d":11,"1.b":"3"}`

### Complex JSON strings

In case of complex JSON structures with arrays, sub-structures, arrays of sub-structures, etc. the indexes are appended at each nesting level:

```json
[
  {
    "UserId": "AIDARRRRRRRRRRRR",
    "UserName": "s3-operator",
    "InlinePolicies": [
      {
        "PolicyName": "policy-s3-operator",
        "Statement": [
          {
            "Effect": "Allow",
            "Action": ["s3:ListAllMyBuckets"],
            "Resource": ["arn:aws:s3:::*"]
          },
          {
            "Effect": "Allow",
            "Action": ["s3:ListBucket", "s3:GetBucketLocation"],
            "Resource": ["arn:aws:s3:::personal-s3-bucket/*"]
          },
          {
            "Effect": "Allow",
            "Action": [
              "s3:PutObject",
              "s3:GetObject",
              "s3:AbortMultipartUpload",
              "s3:ListMultipartUploadParts",
              "s3:ListBucketMultipartUploads"
            ],
            "Resource": ["arn:aws:s3:::personal-s3-bucket/*"]
          }
        ]
      }
    ]
  }
]
```

Output:

```json
{
  "0.UserId": "AIDARRRRRRRRRRRR",
  "0.UserName": "s3-operator",
  "0.InlinePolicies.0.PolicyName": "policy-s3-operator",
  "0.InlinePolicies.0.Statement.0.Effect": "Allow",
  "0.InlinePolicies.0.Statement.0.Action.0": "s3:ListAllMyBuckets",
  "0.InlinePolicies.0.Statement.0.Resource.0": "arn:aws:s3:::*",
  "0.InlinePolicies.0.Statement.1.Effect": "Allow",
  "0.InlinePolicies.0.Statement.1.Action.0": "s3:ListBucket",
  "0.InlinePolicies.0.Statement.1.Action.1": "s3:GetBucketLocation",
  "0.InlinePolicies.0.Statement.1.Resource.0": "arn:aws:s3:::personal-s3-bucket/*",
  "0.InlinePolicies.0.Statement.2.Effect": "Allow",
  "0.InlinePolicies.0.Statement.2.Action.0": "s3:PutObject",
  "0.InlinePolicies.0.Statement.2.Action.1": "s3:GetObject",
  "0.InlinePolicies.0.Statement.2.Action.2": "s3:AbortMultipartUpload",
  "0.InlinePolicies.0.Statement.2.Action.3": "s3:ListMultipartUploadParts",
  "0.InlinePolicies.0.Statement.2.Action.4": "s3:ListBucketMultipartUploads",
  "0.InlinePolicies.0.Statement.2.Resource.0": "arn:aws:s3:::personal-s3-bucket/*"
}
```

### Structs

You can flatten Go structs directly, including those with pointers (common in AWS SDK types):

```golang
package main

import (
    "fmt"

    goflat "github.com/notdodo/goflat/v2"
)

type Instance struct {
    InstanceId   *string
    InstanceType *string
    PublicIp     *string
    Running      *bool
    VCPUs        *int64
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func int64Ptr(i int64) *int64 { return &i }

func main() {
    inst := Instance{
        InstanceId:   strPtr("i-123456"),
        InstanceType: strPtr("t3.micro"),
        PublicIp:     nil,
        Running:      boolPtr(true),
        VCPUs:        int64Ptr(2),
    }

    result := goflat.FlatStruct(inst, goflat.FlattenerConfig{
        Separator: ".",
        OmitEmpty: true,
        OmitNil:   true,
    })
    fmt.Println(result)
    // map[InstanceId:i-123456 InstanceType:t3.micro Running:true VCPUs:2]
    // PublicIp is omitted because it's nil
}
```

If a struct field contains a JSON string (e.g. an AWS inline policy stored as `*string`), goflat will detect and recursively flatten it:

```golang
type Member struct {
    User     *User
    Role     string
    SubField *string // contains JSON
}
```

The JSON inside `SubField` is expanded with dot-notation keys like `Members.0.SubField.0.PolicyName`.
