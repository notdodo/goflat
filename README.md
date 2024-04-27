# goflat

[![Golang CI](https://github.com/notdodo/goflat/actions/workflows/go-ci.yml/badge.svg)](https://github.com/notdodo/goflat/actions/workflows/go-ci.yml)

Flatten complex JSON structures to a one-dimensional map (JSON key/value) that can be converted to a `map[string]interface{}`.

`goflat` supports the flattening of:

- Structs
- JSON strings
- Maps

## Examples

### Using a basic JSON structure

```golang
package main

import (
	"fmt"
	"log"

	"github.com/notdodo/goflat"
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

Output is: `{"a":"3","b.a":"","b.c":true,"b.e":null}`; the sub-structure is returned as a single JSON object.

To also remove the `null` you can pass the struct:

```golang
FlattenerConfig{
		Prefix:    "",
		Separator: ".",
		OmitEmpty: false,
		OmitNil:   true,
		SortKeys:  false,
}
```

### Arrays

When dealing with arrays and recursive structures the library will handle the depth using indexes:

Input: `[{"a": "3"}, {"b": "3", "C": [{"c": 10}, {"d": 11}]}]`

Output: `{"0.a":"3","1.C.0.c":10,"1.C.1.d":11,"1.b":"3"}`

### Complex JSON strings

In case of complex JSON structures with array, sub-structures, arrays of sub-structures, etc the previous statement remains valid but in case of multiple array the indexes are always appended in the last part.

For example:

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

In this case the output is

```json
{
  "0.InlinePolicies.0.Statement.1.Action.0": "s3:ListBucket",
  "0.InlinePolicies.0.Statement.2.Action.4": "s3:ListBucketMultipartUploads",
  "0.InlinePolicies.0.Statement.2.Effect": "Allow",
  "0.InlinePolicies.0.Statement.1.Action.1": "s3:GetBucketLocation",
  "0.InlinePolicies.0.Statement.2.Action.1": "s3:GetObject",
  "0.InlinePolicies.0.Statement.2.Action.3": "s3:ListMultipartUploadParts",
  "0.UserName": "s3-operator",
  "0.InlinePolicies.0.Statement.0.Action.0": "s3:ListAllMyBuckets",
  "0.InlinePolicies.0.Statement.1.Resource.0": "arn:aws:s3:::personal-s3-bucket/*",
  "0.InlinePolicies.0.Statement.0.Effect": "Allow",
  "0.UserId": "AIDARRRRRRRRRRRR",
  "0.InlinePolicies.0.Statement.0.Resource.0": "arn:aws:s3:::*",
  "0.InlinePolicies.0.Statement.1.Effect": "Allow",
  "0.InlinePolicies.0.Statement.2.Action.0": "s3:PutObject",
  "0.InlinePolicies.0.Statement.2.Action.2": "s3:AbortMultipartUpload",
  "0.InlinePolicies.0.PolicyName": "policy-s3-operator",
  "0.InlinePolicies.0.Statement.2.Resource.0": "arn:aws:s3:::personal-s3-bucket/*"
}
```

The `[]map[string]interface{}` can be created using `json.Unmarshal([]byte(myJsonString), &myArrayMapStringInterface)`

### Structs

You can also use the library to flatten any valid struct simply Marshalling the struct to a JSON string

```golang
package main

import (
	"encoding/json"
	"fmt"

	"github.com/notdodo/goflat"
)

type Sub struct {
	A int
	B string
	C []SubSub
}

type SubSub struct {
	D int
}

func main() {
	structOne := Sub{
		A: 3,
		B: "hello",
		C: []SubSub{
			{D: 10}, {D: 11},
		},
	}

	flatten := goflat.FlatStruct(structOne)
	fmt.Println(flatten)
	jsonStr, _ := json.Marshal(flatten)
	fmt.Println(string(jsonStr))
}
```

The output is:

```json
map[A:3 B:hello C..0:{10} C..1:{11}]
{"A":3,"B":"hello","C..0":{"D":10},"C..1":{"D":11}}
```
