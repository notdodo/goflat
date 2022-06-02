# goflat

Flatten complex JSON structures to a one-dimensional map (JSON key/value).

## Examples

Using a basic JSON structure

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

When dealing with arrays and recursive structures the library will handle the depth using indexes:

Input: `[{"a": "3"}, {"b": "3", "C": [{"c": 10}, {"d": 11}]}]`

Output: `[{"a":"3"},{"C.c.0":10,"C.d.1":11,"b":"3"}]`

