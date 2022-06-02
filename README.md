# goflat

Flatten complex JSON structures to a one-dimensional map (JSON key/value) that can be converted to a `map[string]interface{}`.

## Examples

### Using a basic JSON structure

```golang
package main

import (
	"fmt"

	"github.com/notdodo/goflat"
)

func main() {
	fmt.Println(goflat.Flat(`{"a": "3", "b": {"c":true}}`, "", "."))
}
```

Output is: `{"a":"3","b.c":true}`; the sub-structure is returned as a single JSON object.

### Arrays

When dealing with arrays and recursive structures the library will handle the depth using indexes:

Input: `[{"a": "3"}, {"b": "3", "C": [{"c": 10}, {"d": 11}]}]`

Output: `[{"a":"3"},{"C.c.0":10,"C.d.1":11,"b":"3"}]`

### Complex JSON strings

In case of complex JSON structures with array, sub-structures, arrays of sub-structures, etc the previous statement remains valid but in case of multiple array the indexes are always appended in the last part.

For example:

```json
[
   {
      "UserId":"AIDARRRRRRRRRRRR",
      "UserName":"s3-operator",
      "InlinePolicies":[
         {
            "PolicyName":"policy-s3-operator",
            "Statement":[
               {
                  "Effect":"Allow",
                  "Action":[
                     "s3:ListAllMyBuckets"
                  ],
                  "Resource":[
                     "arn:aws:s3:::*"
                  ]
               },
               {
                  "Effect":"Allow",
                  "Action":[
                     "s3:ListBucket",
                     "s3:GetBucketLocation"
                  ],
                  "Resource":[
                     "arn:aws:s3:::personal-s3-bucket/*"
                  ]
               },
               {
                  "Effect":"Allow",
                  "Action":[
                     "s3:PutObject",
                     "s3:GetObject",
                     "s3:AbortMultipartUpload",
                     "s3:ListMultipartUploadParts",
                     "s3:ListBucketMultipartUploads"
                  ],
                  "Resource":[
                     "arn:aws:s3:::personal-s3-bucket/*"
                  ]
               }
            ]
         }
      ]
   }
]
```

In this case the output will be:

```json
[
   {
      "InlinePolicies.PolicyName.0":"policy-s3-operator",
      "InlinePolicies.Statement.Action.0.0.0":"s3:ListAllMyBuckets",
      "InlinePolicies.Statement.Action.0.1.0":"s3:ListBucket",
      "InlinePolicies.Statement.Action.0.2.0":"s3:PutObject",
      "InlinePolicies.Statement.Action.1.1.0":"s3:GetBucketLocation",
      "InlinePolicies.Statement.Action.1.2.0":"s3:GetObject",
      "InlinePolicies.Statement.Action.2.2.0":"s3:AbortMultipartUpload",
      "InlinePolicies.Statement.Action.3.2.0":"s3:ListMultipartUploadParts",
      "InlinePolicies.Statement.Action.4.2.0":"s3:ListBucketMultipartUploads",
      "InlinePolicies.Statement.Effect.0.0":"Allow",
      "InlinePolicies.Statement.Effect.1.0":"Allow",
      "InlinePolicies.Statement.Effect.2.0":"Allow",
      "InlinePolicies.Statement.Resource.0.0.0":"arn:aws:s3:::*",
      "InlinePolicies.Statement.Resource.0.1.0":"arn:aws:s3:::personal-s3-bucket/*",
      "InlinePolicies.Statement.Resource.0.2.0":"arn:aws:s3:::personal-s3-bucket/*",
      "UserId":"AIDARRRRRRRRRRRR",
      "UserName":"s3-operator"
   }
]
```

The `[]map[string]interface{}` can be created using `json.Unmarshal([]byte(myJsonString), &myArrayMapStringInterface)`

The indexes should be read from right to left:

* `InlinePolicies.Statement.Action.4.2.0`: InlinePolicies with index 0 and Statement with index 2 and Action with index 4

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

	jsonStr, _ := json.Marshal(structOne)
	fmt.Println(goflat.Flat(string(jsonStr), "", "."))
}
```

The output is: `{"A":3,"B":"hello","C.D.0":10,"C.D.1":11}`
