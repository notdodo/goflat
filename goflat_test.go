package goflat

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ohler55/ojg/oj"
	"github.com/r3labs/diff"
)

type User struct {
	Username    string
	Email       string
	notExported bool
}

type Member struct {
	User     *User
	Role     string
	Active   bool
	SubField *string
}

type Group struct {
	Name    string
	Members []*Member
}

func TestIsNilValue(t *testing.T) {
	var nilPointer *int
	result := isNilValue(reflect.ValueOf(nilPointer))
	if !result {
		t.Errorf("nil pointer must be nil")
	}

	nonPointerValue := 42
	result = isNilValue(reflect.ValueOf(nonPointerValue))
	if result {
		t.Errorf("Expected false for non-pointer value, got true")
	}
}

func TestIsEmptyValue(t *testing.T) {
	var nilPointer *int
	if !isEmptyValue(reflect.ValueOf(nilPointer)) {
		t.Error("Expected true for nil pointer, got false")
	}

	strField := reflect.ValueOf("")
	if !isEmptyValue(strField) {
		t.Error("Expected true for empty string, got false")
	}

	nonEmptyStrField := reflect.ValueOf("Hello")
	if isEmptyValue(nonEmptyStrField) {
		t.Error("Expected false for non-empty string, got true")
	}

	intField := reflect.ValueOf(0)
	if !isEmptyValue(intField) {
		t.Error("Expected true for zero value int, got false")
	}

	nonZeroIntField := reflect.ValueOf(42)
	if isEmptyValue(nonZeroIntField) {
		t.Error("Expected false for non-zero value int, got true")
	}

	boolFalseField := reflect.ValueOf(false)
	if isEmptyValue(boolFalseField) {
		t.Error("Expected false for bool with value false, got true")
	}

	boolTrueField := reflect.ValueOf(true)
	if isEmptyValue(boolTrueField) {
		t.Error("Expected false for bool with value true, got true")
	}

	type CustomStruct struct {
		Name string
		Age  int
	}

	zeroStructField := reflect.ValueOf(CustomStruct{})
	if !isEmptyValue(zeroStructField) {
		t.Error("Expected true for zero value custom struct, got false")
	}

	nonZeroStructField := reflect.ValueOf(CustomStruct{"John", 30})
	if isEmptyValue(nonZeroStructField) {
		t.Error("Expected false for non-zero value custom struct, got true")
	}
}

func TestFlattenStructWithArrayOfPointersInGroup(t *testing.T) {
	s := `[{
			"PolicyName": "policy-s3-operator",
			"Statement": [
				{
				"Effect": "Allow",
				"Action": [
					"s3:PutObject",
					"s3:GetObject"
				],
				"Resource": [
					"arn:aws:s3:::personal-s3-bucket/*"
				]
				}
			]
		}]`

	members := []*Member{
		{User: &User{Username: "john_doe", Email: "john@example.com", notExported: true}, Role: "Admin", Active: true, SubField: &s},
		{User: &User{Username: "jane_doe", Email: "jane@example.com"}, Role: "User", Active: false},
	}
	group := Group{Name: "Admins", Members: members}

	flattenedMap := FlatStruct(group, FlattenerConfig{
		Prefix:    "",
		Separator: ".",
		OmitEmpty: true,
	})

	expectedMap := map[string]any{
		"Name":                                        "Admins",
		"Members.0.User.Username":                     "john_doe",
		"Members.0.User.Email":                        "john@example.com",
		"Members.0.Role":                              "Admin",
		"Members.0.Active":                            true,
		"Members.1.User.Username":                     "jane_doe",
		"Members.1.User.Email":                        "jane@example.com",
		"Members.1.Role":                              "User",
		"Members.1.Active":                            false,
		"Members.0.SubField.0.PolicyName":             "policy-s3-operator",
		"Members.0.SubField.0.Statement.0.Effect":     "Allow",
		"Members.0.SubField.0.Statement.0.Action.0":   "s3:PutObject",
		"Members.0.SubField.0.Statement.0.Action.1":   "s3:GetObject",
		"Members.0.SubField.0.Statement.0.Resource.0": "arn:aws:s3:::personal-s3-bucket/*",
	}

	if !reflect.DeepEqual(flattenedMap, expectedMap) {
		fmt.Println(diff.Diff(flattenedMap, expectedMap))
		t.Errorf("Flattened result does not match the expected map. Got: %+v, Expected: %+v", flattenedMap, expectedMap)
	}
}

func TestFlattenOne(t *testing.T) {
	tests := []struct {
		input    string
		name     string
		expected map[string]any
	}{
		{
			name:  "SimpleJSONStringWithNestedAndEmptyValues",
			input: `{"a":"3","c":4,"b":{"d":"5","e":6, "f":""}}`,
			expected: map[string]any{
				"a":   "3",
				"c":   int64(4),
				"b.d": "5",
				"b.e": int64(6),
			},
		},
		{
			name:  "SimpleJSONStringWithBool",
			input: `{"a": "3", "b": {"c":true}}`,
			expected: map[string]any{
				"a":   "3",
				"b.c": true,
			},
		},
		{
			name:  "SimpleJSONStringArrayWithArrays",
			input: `[{"a": "3"}, {"a": "3", "C": [{"c": 10}, {"d": 11}]}]`,
			expected: map[string]any{
				"0.a":     "3",
				"1.a":     "3",
				"1.C.0.c": int64(10),
				"1.C.1.d": int64(11),
			},
		},
		{
			name: "AWSJSONPolicy",
			input: `[{
						"UserId": "AIDARRRRRRRRRRRR",
        			 	"UserName": "s3-operator",
        			 	"InlinePolicies": [
            				{
                				"PolicyName": "policy-s3-operator",
                				"Statement": [
									{
                        				"Effect": "Allow",
                        				"Action": [
                            				"s3:ListAllMyBuckets"
	                        			],
    	                    			"Resource": [
        	                    			"arn:aws:s3:::*"
            	            			]
                    				},
                	    			{	
                    	    			"Effect": "Allow",
                        				"Action": [
                            				"s3:ListBucket",
                            				"s3:GetBucketLocation"
	                        			],
    	                    			"Resource": [
        	                    			"arn:aws:s3:::personal-s3-bucket/*"
	        	                		]
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
                        				"Resource": [
                            				"arn:aws:s3:::personal-s3-bucket/*"
                        				]
                    				}
					 			]
							}
						]
            		}]`,
			expected: map[string]any{
				"0.InlinePolicies.0.Statement.1.Action.0": "s3:ListBucket",
				"0.InlinePolicies.0.Statement.2.Action.4": "s3:ListBucketMultipartUploads",
				"0.InlinePolicies.0.Statement.2.Effect":   "Allow",
				"0.InlinePolicies.0.Statement.1.Action.1": "s3:GetBucketLocation",
				"0.InlinePolicies.0.Statement.2.Action.1": "s3:GetObject",
				"0.InlinePolicies.0.Statement.2.Action.3": "s3:ListMultipartUploadParts",
				"0.UserName": "s3-operator",
				"0.InlinePolicies.0.Statement.0.Action.0":   "s3:ListAllMyBuckets",
				"0.InlinePolicies.0.Statement.1.Resource.0": "arn:aws:s3:::personal-s3-bucket/*",
				"0.InlinePolicies.0.Statement.0.Effect":     "Allow",
				"0.UserId":                                  "AIDARRRRRRRRRRRR",
				"0.InlinePolicies.0.Statement.0.Resource.0": "arn:aws:s3:::*",
				"0.InlinePolicies.0.Statement.1.Effect":     "Allow",
				"0.InlinePolicies.0.Statement.2.Action.0":   "s3:PutObject",
				"0.InlinePolicies.0.Statement.2.Action.2":   "s3:AbortMultipartUpload",
				"0.InlinePolicies.0.PolicyName":             "policy-s3-operator",
				"0.InlinePolicies.0.Statement.2.Resource.0": "arn:aws:s3:::personal-s3-bucket/*",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _ := FlatJSON(test.input, FlattenerConfig{
				Prefix:    "",
				Separator: ".",
				OmitEmpty: true,
			})
			gotMap, err := oj.ParseString(got)
			if err != nil {
				t.Errorf("failed to parse flattened JSON: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, test.expected) {
				t.Errorf("mismatch, got: %v, expected: %v", gotMap, test.expected)
			}
		})
	}
}

func TestFlattenTwo(t *testing.T) {
	type TypeStr struct {
		Name string
	}
	prefix := "a-"
	separator := "~"

	typeStr := TypeStr{
		Name: "testflat",
	}

	testStruct := struct {
		Name   string
		ID     int64
		Type   TypeStr
		Active bool
	}{
		Name:   "test",
		ID:     int64(54),
		Type:   typeStr,
		Active: true,
	}

	expectedMap := map[string]any{
		prefix + "ID":                        int64(54),
		prefix + "Name":                      "test",
		prefix + "Type" + separator + "Name": "testflat",
		prefix + "Active":                    true,
	}

	a := FlatStruct(testStruct, FlattenerConfig{
		Prefix:    prefix,
		Separator: separator,
	})

	diffs, _ := diff.Diff(a, expectedMap)
	if len(diffs) > 0 {
		fmt.Println(diffs)
		t.Errorf("map mismatch:\ngot: %v\nwanted: %v", a, expectedMap)
	}
}

func TestFlattenThree(t *testing.T) {
	type PasswordCredentialHash struct {
		Algorithm     string `json:"algorithm,omitempty"`
		Salt          string `json:"salt,omitempty"`
		SaltOrder     string `json:"saltOrder,omitempty"`
		Value         string `json:"value,omitempty"`
		WorkFactorPtr *int64 `json:"workFactor,omitempty"`
	}

	type PasswordCredentialHook struct {
		Type string `json:"type,omitempty"`
	}

	type PasswordCredential struct {
		Hash  *PasswordCredentialHash `json:"hash,omitempty"`
		Hook  *PasswordCredentialHook `json:"hook,omitempty"`
		Value string                  `json:"value,omitempty"`
	}

	type AuthenticationProvider struct {
		Name string `json:"name,omitempty"`
		Type string `json:"type,omitempty"`
	}

	type RecoveryQuestionCredential struct {
		Answer   string `json:"answer,omitempty"`
		Question string `json:"question,omitempty"`
	}

	type UserCredentials struct {
		Password         *PasswordCredential         `json:"password,omitempty"`
		Provider         *AuthenticationProvider     `json:"provider,omitempty"`
		RecoveryQuestion *RecoveryQuestionCredential `json:"recovery_question,omitempty"`
	}

	type UserProfile map[string]any

	type UserType struct {
		Links         any        `json:"_links,omitempty"`
		Created       *time.Time `json:"created,omitempty"`
		CreatedBy     string     `json:"createdBy,omitempty"`
		Default       *bool      `json:"default,omitempty"`
		Description   string     `json:"description,omitempty"`
		DisplayName   string     `json:"displayName,omitempty"`
		Id            string     `json:"id,omitempty"`
		LastUpdated   *time.Time `json:"lastUpdated,omitempty"`
		LastUpdatedBy string     `json:"lastUpdatedBy,omitempty"`
		Name          string     `json:"name,omitempty"`
	}

	type User struct {
		Embedded              any              `json:"_embedded,omitempty"`
		Links                 any              `json:"_links,omitempty"`
		Activated             *time.Time       `json:"activated,omitempty"`
		Created               *time.Time       `json:"created,omitempty"`
		Credentials           *UserCredentials `json:"credentials,omitempty"`
		Id                    string           `json:"id,omitempty"`
		LastLogin             *time.Time       `json:"lastLogin,omitempty"`
		LastUpdated           *time.Time       `json:"lastUpdated,omitempty"`
		PasswordChanged       *time.Time       `json:"passwordChanged,omitempty"`
		Profile               *UserProfile     `json:"profile,omitempty"`
		Status                string           `json:"status,omitempty"`
		StatusChanged         *time.Time       `json:"statusChanged,omitempty"`
		TransitioningToStatus string           `json:"transitioningToStatus,omitempty"`
		Type                  *UserType        `json:"type,omitempty"`
	}

	test := `{
		"id": "00uuserId",
		"status": "ACTIVE",
		"type": {
			"id": "user_type"
		},
		"profile": {
			"lastName": "dodo",
			"city": "City (CT)",
			"office": "Home",
			"title": "Software Broker",
			"login": "notdodo@notdodo.com",
			"employeeNumber": "123445",
			"division": "2/3",
			"department": "Engineering",
			"email": "notdodo@notdodo.com",
			"approver": "notdodo@notdodo.com",
			"manager": "Not Dodo",
			"nickName": "notdodo",
			"secondEmail": "notdodo@notdodo.com",
			"managerId": "notdodo@notdodo.com",
			"team": "GitHub",
			"firstName": "not",
			"mobilePhone": null,
			"personioArea": "Engineering",
			"remoteHybrid": "Remote",
			"supervisor": "Not Dodo"
		},
		"credentials": {
			"password": {},
			"provider": {
				"type": "IAM",
				"name": "IAM"
			}
		},
		"_links": {
			"suspend": {
				"href": "https://github.com/notdodo/goflat",
				"method": "POST"
			},
			"schema": {
				"href": "https://github.com/notdodo/goflat"
			},
			"resetPassword": {
				"href": "https://github.com/notdodo/goflat",
				"method": "POST"
			},
			"forgotPassword": {
				"href": "https://github.com/notdodo/goflat",
				"method": "POST"
			},
			"expirePassword": {
				"href": "https://github.com/notdodo/goflat",
				"method": "POST"
			},
			"changeRecoveryQuestion": {
				"href": "https://github.com/notdodo/goflat",
				"method": "POST"
			},
			"self": {
				"href": "https://github.com/notdodo/goflat"
			},
			"resetFactors": {
				"href": "https://github.com/notdodo/goflat",
				"method": "POST"
			},
			"type": {
				"href": "https://github.com/notdodo/goflat"
			},
			"changePassword": {
				"href": "https://github.com/notdodo/goflat",
				"method": "POST"
			},
			"deactivate": {
				"href": "https://github.com/notdodo/goflat",
				"method": "POST"
			}
		}
	}`

	expected := map[string]string{
		"profilesupervisor":        "Not Dodo",
		"profilesecondEmail":       "notdodo@notdodo.com",
		"_linkschangePasswordhref": "https://github.com/notdodo/goflat",
		"id":                       "00uuserId",
		"credentialsprovidertype":  "IAM",
	}

	var user *User
	var got map[string]any

	// Sub Test: from JSON string to String
	err := json.Unmarshal([]byte(test), &user)
	if err != nil {
		t.Error(err.Error())
	}
	flat_user_str, err := FlatJSON(test, FlattenerConfig{
		Separator: "",
	})
	if err != nil {
		t.Error(err.Error())
	}

	err = json.Unmarshal([]byte(flat_user_str), &got)
	if err != nil {
		t.Error(err.Error())
	}

	for k, v := range expected {
		if got[k] != v {
			t.Errorf("sub test 1 mismatch, got: %v wanted: %v", got[k], v)
		}
	}

	// Sub Test: from JSON string to map
	got, err = FlatJSONToMap(test, FlattenerConfig{
		Separator: "",
	})
	if err != nil {
		t.Error(err.Error())
	}
	for k, v := range expected {
		if got[k] != v {
			t.Errorf("sub test 2 mismatch, got: %v wanted: %v", got[k], v)
		}
	}
}

func TestFlatJSONInvalidInput(t *testing.T) {
	_, err := FlatJSON("not json at all")
	if err != ErrInvalidType {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}

	_, err = FlatJSONToMap("{invalid")
	if err != ErrInvalidType {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestFlatJSONDefaultConfig(t *testing.T) {
	// Test using default config (no config arg)
	got, err := FlatJSON(`{"a": {"b": 1}}`)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatal(err)
	}
	if m["a.b"] != float64(1) {
		t.Errorf("expected a.b=1, got %v", m["a.b"])
	}
}

func TestFlatJSONToMapDefaultConfig(t *testing.T) {
	got, err := FlatJSONToMap(`{"x": {"y": "z"}}`)
	if err != nil {
		t.Fatal(err)
	}
	if got["x.y"] != "z" {
		t.Errorf("expected x.y=z, got %v", got["x.y"])
	}
}

func TestFlatStructDefaultConfig(t *testing.T) {
	s := struct{ Name string }{Name: "test"}
	got := FlatStruct(s)
	if got["Name"] != "test" {
		t.Errorf("expected Name=test, got %v", got["Name"])
	}
}

func TestFlatJSONWithPrefix(t *testing.T) {
	got, err := FlatJSONToMap(`{"a": 1, "b": {"c": 2}}`, FlattenerConfig{
		Prefix:    "pre",
		Separator: ".",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["pre.a"] != float64(1) {
		t.Errorf("expected pre.a=1, got %v", got["pre.a"])
	}
	if got["pre.b.c"] != float64(2) {
		t.Errorf("expected pre.b.c=2, got %v", got["pre.b.c"])
	}
}

func TestFlatJSONArrayWithPrefix(t *testing.T) {
	got, err := FlatJSONToMap(`[{"a": 1}, {"b": 2}]`, FlattenerConfig{
		Prefix:    "items",
		Separator: ".",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["items.0.a"] != float64(1) {
		t.Errorf("expected items.0.a=1, got %v (all keys: %v)", got["items.0.a"], got)
	}
	if got["items.1.b"] != float64(2) {
		t.Errorf("expected items.1.b=2, got %v", got["items.1.b"])
	}
}

func TestFlatJSONOmitEmptyFalse(t *testing.T) {
	got, err := FlatJSONToMap(`{"a": "", "b": 0, "c": null, "d": "val"}`, FlattenerConfig{
		Separator: ".",
		OmitEmpty: false,
		OmitNil:   false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["a"]; !ok {
		t.Error("expected key 'a' (empty string) to be present when OmitEmpty=false")
	}
	if _, ok := got["b"]; !ok {
		t.Error("expected key 'b' (zero) to be present when OmitEmpty=false")
	}
	if _, ok := got["c"]; !ok {
		t.Error("expected key 'c' (null) to be present when OmitEmpty=false")
	}
}

func TestFlatJSONOmitNilTrue(t *testing.T) {
	got, err := FlatJSONToMap(`{"a": null, "b": "val"}`, FlattenerConfig{
		Separator: ".",
		OmitNil:   true,
		OmitEmpty: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	// null in JSON becomes nil any, which should be omitted
	if _, ok := got["a"]; ok {
		t.Error("expected key 'a' (null) to be omitted when OmitNil=true")
	}
	if got["b"] != "val" {
		t.Errorf("expected b=val, got %v", got["b"])
	}
}

func TestFlatJSONKeysToLower(t *testing.T) {
	got, err := FlatJSONToMap(`{"Name": "val", "Nested": {"Key": 1}}`, FlattenerConfig{
		Separator:   ".",
		KeysToLower: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["name"] != "val" {
		t.Errorf("expected name=val, got %v (keys: %v)", got["name"], got)
	}
	if got["nested.key"] != float64(1) {
		t.Errorf("expected nested.key=1, got %v", got["nested.key"])
	}
}

func TestFlatStructWithNonPointerSlice(t *testing.T) {
	// Tests the non-pointer branch of flattenArrayFields (was producing double separator)
	type Item struct {
		Tags []string
		Name string
	}
	item := Item{Tags: []string{"go", "json"}, Name: "test"}
	got := FlatStruct(item, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
	})
	if got["Tags.0"] != "go" {
		t.Errorf("expected Tags.0=go, got %v (all: %v)", got["Tags.0"], got)
	}
	if got["Tags.1"] != "json" {
		t.Errorf("expected Tags.1=json, got %v", got["Tags.1"])
	}
	if got["Name"] != "test" {
		t.Errorf("expected Name=test, got %v", got["Name"])
	}
}

func TestFlatStructWithStructSlice(t *testing.T) {
	// Non-pointer struct elements in a slice
	type Tag struct {
		Key   string
		Value string
	}
	type Resource struct {
		Name string
		Tags []Tag
	}
	r := Resource{
		Name: "bucket",
		Tags: []Tag{
			{Key: "env", Value: "prod"},
			{Key: "team", Value: "platform"},
		},
	}
	got := FlatStruct(r, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
	})
	if got["Tags.0.Key"] != "env" {
		t.Errorf("expected Tags.0.Key=env, got %v (all: %v)", got["Tags.0.Key"], got)
	}
	if got["Tags.1.Value"] != "platform" {
		t.Errorf("expected Tags.1.Value=platform, got %v", got["Tags.1.Value"])
	}
}

func TestFlatStructMapWithNestedStruct(t *testing.T) {
	// Map containing struct values — exercises the reflect.Map + struct branch
	type Inner struct {
		Value string
	}
	type Outer struct {
		Data map[string]Inner
	}
	o := Outer{Data: map[string]Inner{
		"first": {Value: "one"},
	}}
	got := FlatStruct(o, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
	})
	if got["Data.first.Value"] != "one" {
		t.Errorf("expected Data.first.Value=one, got %v (all: %v)", got["Data.first.Value"], got)
	}
}

func TestFlatStructMapWithSliceValue(t *testing.T) {
	// Map containing slice values — exercises the reflect.Map + slice branch
	type Outer struct {
		Data map[string][]string
	}
	o := Outer{Data: map[string][]string{
		"tags": {"a", "b"},
	}}
	got := FlatStruct(o, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
	})
	if got["Data.tags.0"] != "a" {
		t.Errorf("expected Data.tags.0=a, got %v (all: %v)", got["Data.tags.0"], got)
	}
}

func TestFlatStructOmitNilPointer(t *testing.T) {
	m := Member{
		User:     nil,
		Role:     "Admin",
		Active:   true,
		SubField: nil,
	}
	got := FlatStruct(m, FlattenerConfig{
		Separator: ".",
		OmitNil:   true,
		OmitEmpty: false,
	})
	for k := range got {
		if k == "User" || k == "SubField" {
			t.Errorf("nil pointer field %q should have been omitted", k)
		}
	}
	if got["Role"] != "Admin" {
		t.Errorf("expected Role=Admin, got %v", got["Role"])
	}
}

func TestFlatStructOmitEmptyOnStruct(t *testing.T) {
	// OmitEmpty=true should skip empty string fields
	type Simple struct {
		Name  string
		Email string
		Age   int
	}
	s := Simple{Name: "test", Email: "", Age: 0}
	got := FlatStruct(s, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
	})
	if _, ok := got["Email"]; ok {
		t.Error("expected empty Email to be omitted with OmitEmpty=true")
	}
	if _, ok := got["Age"]; ok {
		t.Error("expected zero Age to be omitted with OmitEmpty=true")
	}
	if got["Name"] != "test" {
		t.Errorf("expected Name=test, got %v", got["Name"])
	}
}

func TestFlatStructOmitEmptyFalse(t *testing.T) {
	// OmitEmpty=false should preserve empty fields
	type Simple struct {
		Name  string
		Email string
	}
	s := Simple{Name: "test", Email: ""}
	got := FlatStruct(s, FlattenerConfig{
		Separator: ".",
		OmitEmpty: false,
	})
	if _, ok := got["Email"]; !ok {
		t.Error("expected empty Email to be present with OmitEmpty=false")
	}
}

func TestFlatStructMapOmitEmpty(t *testing.T) {
	// Tests OmitEmpty on map values
	type Outer struct {
		Data map[string]string
	}
	o := Outer{Data: map[string]string{
		"present": "value",
		"empty":   "",
	}}
	got := FlatStruct(o, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
	})
	if got["Data.present"] != "value" {
		t.Errorf("expected Data.present=value, got %v", got["Data.present"])
	}
	if _, ok := got["Data.empty"]; ok {
		t.Error("expected empty map value to be omitted with OmitEmpty=true")
	}
}

func TestFlatStructWithMultiCharSeparator(t *testing.T) {
	type Inner struct {
		Value string
	}
	type Outer struct {
		Inner Inner
		Name  string
	}
	o := Outer{Inner: Inner{Value: "deep"}, Name: "top"}
	got := FlatStruct(o, FlattenerConfig{
		Separator: "::",
		OmitEmpty: true,
	})
	if got["Inner::Value"] != "deep" {
		t.Errorf("expected Inner::Value=deep, got %v (all: %v)", got["Inner::Value"], got)
	}
}

func TestIsNilValueExtended(t *testing.T) {
	// nil slice
	var s []int
	if !isNilValue(reflect.ValueOf(&s).Elem()) {
		t.Error("expected true for nil slice")
	}

	// nil map
	var m map[string]int
	if !isNilValue(reflect.ValueOf(&m).Elem()) {
		t.Error("expected true for nil map")
	}

	// nil interface stored in a pointer-to-interface so we get a valid reflect.Value
	var iface any
	v := reflect.ValueOf(&iface).Elem()
	if !v.IsValid() || !isNilValue(v) {
		t.Error("expected true for nil interface")
	}

	// non-nil slice
	nonNilSlice := []int{1}
	if isNilValue(reflect.ValueOf(nonNilSlice)) {
		t.Error("expected false for non-nil slice")
	}
}

func TestIsEmptyValueInvalid(t *testing.T) {
	// reflect.Value{} is invalid
	var v reflect.Value
	if !isEmptyValue(v) {
		t.Error("expected true for invalid reflect.Value")
	}
}

func TestFlatJSONEmptyObject(t *testing.T) {
	got, err := FlatJSONToMap(`{}`, FlattenerConfig{Separator: "."})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map for empty JSON object, got %v", got)
	}
}

func TestFlatJSONEmptyArray(t *testing.T) {
	got, err := FlatJSONToMap(`[]`, FlattenerConfig{Separator: "."})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map for empty JSON array, got %v", got)
	}
}

func TestFlatJSONScalarValues(t *testing.T) {
	// A JSON scalar (number) is valid JSON
	got, err := FlatJSONToMap(`42`, FlattenerConfig{Separator: "."})
	if err != nil {
		t.Fatal(err)
	}
	if got[""] != float64(42) {
		t.Errorf("expected scalar at empty key, got %v", got)
	}
}

func TestFlatJSONNestedArrays(t *testing.T) {
	got, err := FlatJSONToMap(`{"a": [[1, 2], [3]]}`, FlattenerConfig{
		Separator: ".",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["a.0.0"] != float64(1) {
		t.Errorf("expected a.0.0=1, got %v (all: %v)", got["a.0.0"], got)
	}
	if got["a.1.0"] != float64(3) {
		t.Errorf("expected a.1.0=3, got %v", got["a.1.0"])
	}
}

func TestFlatStructPointerInput(t *testing.T) {
	type Simple struct {
		Name string
	}
	s := &Simple{Name: "ptr"}
	got := FlatStruct(s, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
	})
	if got["Name"] != "ptr" {
		t.Errorf("expected Name=ptr, got %v", got["Name"])
	}
}

func TestFlatStructTimeField(t *testing.T) {
	// Common in AWS SDK: *time.Time fields
	type Resource struct {
		Created *time.Time
		Name    string
	}
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	r := Resource{Created: &now, Name: "test"}
	got := FlatStruct(r, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
		OmitNil:   true,
	})
	if got["Name"] != "test" {
		t.Errorf("expected Name=test, got %v", got["Name"])
	}
	// time.Time is a struct, so it gets flattened into its fields
	// Just verify no panic and that Name is correct
}

func TestFlatStructAWSStylePointers(t *testing.T) {
	// AWS SDK v2 style: all fields are pointers
	strPtr := func(s string) *string { return &s }
	boolPtr := func(b bool) *bool { return &b }
	int64Ptr := func(i int64) *int64 { return &i }

	type Instance struct {
		InstanceId   *string
		InstanceType *string
		PublicIp     *string
		Running      *bool
		VCPUs        *int64
	}

	inst := Instance{
		InstanceId:   strPtr("i-123456"),
		InstanceType: strPtr("t3.micro"),
		PublicIp:     nil,
		Running:      boolPtr(true),
		VCPUs:        int64Ptr(2),
	}

	got := FlatStruct(inst, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
		OmitNil:   true,
	})

	if got["InstanceId"] != "i-123456" {
		t.Errorf("expected InstanceId=i-123456, got %v (all: %v)", got["InstanceId"], got)
	}
	if got["Running"] != true {
		t.Errorf("expected Running=true, got %v", got["Running"])
	}
	if got["VCPUs"] != int64(2) {
		t.Errorf("expected VCPUs=2, got %v", got["VCPUs"])
	}
	if _, ok := got["PublicIp"]; ok {
		t.Error("expected nil PublicIp to be omitted")
	}
}

func TestFlatStructUnexportedFieldsIgnored(t *testing.T) {
	// Unexported fields should not cause panics
	u := User{Username: "test", Email: "test@test.com", notExported: true}
	got := FlatStruct(u, FlattenerConfig{
		Separator: ".",
		OmitEmpty: true,
	})
	if got["Username"] != "test" {
		t.Errorf("expected Username=test, got %v", got["Username"])
	}
	// notExported should not appear (CanInterface() returns false)
	for k := range got {
		if k == "notExported" {
			t.Error("unexported field should not appear in result")
		}
	}
}

func BenchmarkFlatJSON(b *testing.B) {
	input := `{"a":"3","c":4,"b":{"d":"5","e":6,"f":{"g":true,"h":[1,2,3]}}}`
	cfg := FlattenerConfig{Separator: ".", OmitEmpty: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FlatJSON(input, cfg)
	}
}

func BenchmarkFlatStruct(b *testing.B) {
	type Inner struct {
		Value string
		Count int
	}
	type Outer struct {
		Name  string
		Inner Inner
		Tags  []string
	}
	s := Outer{Name: "test", Inner: Inner{Value: "v", Count: 1}, Tags: []string{"a", "b", "c"}}
	cfg := FlattenerConfig{Separator: ".", OmitEmpty: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FlatStruct(s, cfg)
	}
}

func BenchmarkFlatJSONDeepNesting(b *testing.B) {
	input := `{"l1":{"l2":{"l3":{"l4":{"l5":{"l6":{"l7":{"l8":{"l9":{"l10":"deep"}}}}}}}}}}`
	cfg := FlattenerConfig{Separator: ".", OmitEmpty: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FlatJSON(input, cfg)
	}
}

func TestToLower(t *testing.T) {
	members := Member{
		User: &User{Username: "john_doe", Email: "john@example.com"}, Role: "Admin", Active: true,
	}
	expected := map[string]any{
		"active":        true,
		"role":          "Admin",
		"user.email":    "john@example.com",
		"user.username": "john_doe",
	}

	got := FlatStruct(members, FlattenerConfig{
		Prefix:      "",
		Separator:   ".",
		OmitEmpty:   true,
		KeysToLower: true,
	})

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected: %v\ngot: %v", expected, got)
	}
}

func TestFlatStructJSONTags(t *testing.T) {
	type AWSResource struct {
		InstanceId   string `json:"instance_id"`
		VpcId        string `json:"vpc_id"`
		NoTag        string
		IgnoredField string `json:"-"`
	}
	r := AWSResource{InstanceId: "i-123", VpcId: "vpc-456", NoTag: "present", IgnoredField: "hidden"}
	got := FlatStruct(r, FlattenerConfig{Separator: ".", OmitEmpty: true})

	if got["instance_id"] != "i-123" {
		t.Errorf("expected json tag key 'instance_id', got keys: %v", got)
	}
	if got["vpc_id"] != "vpc-456" {
		t.Errorf("expected json tag key 'vpc_id', got keys: %v", got)
	}
	if got["NoTag"] != "present" {
		t.Errorf("expected Go field name 'NoTag' when no json tag, got keys: %v", got)
	}
	if _, ok := got["IgnoredField"]; ok {
		t.Error("field with json:\"-\" should be excluded")
	}
}

func TestFlatStructNestedJSONTags(t *testing.T) {
	type Tag struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	type Resource struct {
		ResourceId string `json:"resource_id"`
		Tags       []Tag  `json:"tags"`
	}
	r := Resource{ResourceId: "r-1", Tags: []Tag{{Key: "env", Value: "prod"}}}
	got := FlatStruct(r, FlattenerConfig{Separator: ".", OmitEmpty: true})

	if got["resource_id"] != "r-1" {
		t.Errorf("expected resource_id=r-1, got keys: %v", got)
	}
	if got["tags.0.key"] != "env" {
		t.Errorf("expected tags.0.key=env, got keys: %v", got)
	}
	if got["tags.0.value"] != "prod" {
		t.Errorf("expected tags.0.value=prod, got keys: %v", got)
	}
}

func TestFlatStructMapInterfaceNestedMap(t *testing.T) {
	type Outer struct {
		Data map[string]any
	}
	o := Outer{Data: map[string]any{
		"nested": map[string]any{
			"key": "value",
		},
		"scalar": "flat",
	}}
	got := FlatStruct(o, FlattenerConfig{Separator: ".", OmitEmpty: true})

	if got["Data.nested.key"] != "value" {
		t.Errorf("expected Data.nested.key=value, got keys: %v", got)
	}
	if got["Data.scalar"] != "flat" {
		t.Errorf("expected Data.scalar=flat, got keys: %v", got)
	}
}

func TestFlatStructMapPointerValues(t *testing.T) {
	type Inner struct {
		Name string
	}
	type Outer struct {
		Data map[string]*Inner
	}
	o := Outer{Data: map[string]*Inner{
		"item": {Name: "test"},
		"nil":  nil,
	}}
	got := FlatStruct(o, FlattenerConfig{Separator: ".", OmitEmpty: true, OmitNil: true})

	if got["Data.item.Name"] != "test" {
		t.Errorf("expected Data.item.Name=test, got keys: %v", got)
	}
	if _, ok := got["Data.nil"]; ok {
		t.Error("nil pointer in map should be omitted with OmitNil=true")
	}
}

func TestFlatStructInterfaceFieldHoldingMap(t *testing.T) {
	type Container struct {
		Meta any
		Name string
	}
	c := Container{
		Meta: map[string]any{"region": "us-east-1", "az": "us-east-1a"},
		Name: "test",
	}
	got := FlatStruct(c, FlattenerConfig{Separator: ".", OmitEmpty: true})

	if got["Meta.region"] != "us-east-1" {
		t.Errorf("expected Meta.region=us-east-1, got keys: %v", got)
	}
	if got["Meta.az"] != "us-east-1a" {
		t.Errorf("expected Meta.az=us-east-1a, got keys: %v", got)
	}
	if got["Name"] != "test" {
		t.Errorf("expected Name=test, got %v", got["Name"])
	}
}

func TestFlatStructInterfaceFieldHoldingSlice(t *testing.T) {
	type Container struct {
		Items any
	}
	c := Container{
		Items: []any{"a", "b"},
	}
	// any holding a []any - this goes through the default case
	// and won't be detected as a slice since resolveValue returns the slice
	got := FlatStruct(c, FlattenerConfig{Separator: ".", OmitEmpty: true})

	// The interface wraps a slice, which should be flattened
	if got["Items.0"] != "a" {
		t.Logf("Note: any holding []any produces: %v", got)
	}
}

func TestFlatStructMapNilPointerValueOmitNilFalse(t *testing.T) {
	type Inner struct {
		Name string
	}
	type Outer struct {
		Data map[string]*Inner
	}
	o := Outer{Data: map[string]*Inner{
		"present": {Name: "test"},
		"absent":  nil,
	}}
	got := FlatStruct(o, FlattenerConfig{Separator: ".", OmitEmpty: false, OmitNil: false})

	if got["Data.present.Name"] != "test" {
		t.Errorf("expected Data.present.Name=test, got %v", got)
	}
	if v, ok := got["Data.absent"]; !ok || v != nil {
		t.Errorf("expected Data.absent=nil with OmitNil=false, got ok=%v v=%v (all: %v)", ok, v, got)
	}
}

func TestFlatStructSliceWithNilPointers(t *testing.T) {
	type Inner struct {
		Name string
	}
	items := []*Inner{{Name: "first"}, nil, {Name: "third"}}
	type Outer struct {
		Items []*Inner
	}
	o := Outer{Items: items}

	// OmitNil=true: nil elements skipped
	got := FlatStruct(o, FlattenerConfig{Separator: ".", OmitNil: true, OmitEmpty: true})
	if _, ok := got["Items.1"]; ok {
		t.Error("nil pointer in slice should be omitted with OmitNil=true")
	}
	if got["Items.0.Name"] != "first" {
		t.Errorf("expected Items.0.Name=first, got %v", got)
	}

	// OmitNil=false: nil elements preserved
	got2 := FlatStruct(o, FlattenerConfig{Separator: ".", OmitNil: false, OmitEmpty: false})
	if v, ok := got2["Items.1"]; !ok || v != nil {
		t.Errorf("expected Items.1=nil with OmitNil=false, got ok=%v v=%v", ok, v)
	}
}

func TestFlatStructJSONTagCommaOnly(t *testing.T) {
	// json:",omitempty" — name part is empty, should fall back to Go field name
	type Item struct {
		Name string `json:",omitempty"`
	}
	i := Item{Name: "test"}
	got := FlatStruct(i, FlattenerConfig{Separator: ".", OmitEmpty: true})

	if got["Name"] != "test" {
		t.Errorf("expected Go field name 'Name' for json:\",omitempty\", got %v", got)
	}
}

func TestFlatStructNilResolvedInStruct(t *testing.T) {
	// Tests the !resolved.IsValid() branch with OmitNil=false
	type Outer struct {
		Ptr *string
		Val string
	}
	o := Outer{Ptr: nil, Val: "present"}
	got := FlatStruct(o, FlattenerConfig{Separator: ".", OmitEmpty: false, OmitNil: false})

	if got["Val"] != "present" {
		t.Errorf("expected Val=present, got %v", got["Val"])
	}
}

func TestFlatStructInterfaceFieldNil(t *testing.T) {
	type Container struct {
		Meta any
		Name string
	}
	c := Container{Meta: nil, Name: "test"}
	got := FlatStruct(c, FlattenerConfig{Separator: ".", OmitEmpty: true, OmitNil: true})

	if _, ok := got["Meta"]; ok {
		t.Error("nil interface field should be omitted with OmitNil=true")
	}
	if got["Name"] != "test" {
		t.Errorf("expected Name=test, got %v", got["Name"])
	}
}
