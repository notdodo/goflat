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

func TestFlattenOne(t *testing.T) {
	tests := []struct {
		test     string
		expected map[string]interface{}
	}{
		{
			test: `{"a":"3","c":4,"b":{"d":"5","e":6, "f":""}}`,
			expected: map[string]interface{}{
				"a":   "3",
				"c":   int64(4),
				"b.d": "5",
				"b.e": int64(6),
			},
		},
		{
			test: `{"a": "3", "b": {"c":true}}`,
			expected: map[string]interface{}{
				"a":   "3",
				"b.c": true,
			},
		},
		{
			test: `[{"a": "3"}, {"a": "3", "C": [{"c": 10}, {"d": 11}]}]`,
			expected: map[string]interface{}{
				"0.a":     "3",
				"1.a":     "3",
				"1.C.0.c": int64(10),
				"1.C.1.d": int64(11),
			},
		},
		{
			test: `[{
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
			expected: map[string]interface{}{
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
		got, _ := FlatJSON(test.test, FlattenerConfig{
			Prefix:    "",
			Separator: ".",
			SortKeys:  true,
			OmitEmpty: true,
		})
		gotMap, err := oj.ParseString(got)
		if err != nil {
			t.Errorf("[X] Test failed with error: %v", err)
			continue
		}

		if !reflect.DeepEqual(gotMap, test.expected) {
			fmt.Println(diff.Diff(gotMap, test.expected))
			t.Errorf("mismatch, got: %v wanted: %v", gotMap, test.expected)
		}
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

	expectedMap := map[string]interface{}{
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

	type UserProfile map[string]interface{}

	type UserType struct {
		Links         interface{} `json:"_links,omitempty"`
		Created       *time.Time  `json:"created,omitempty"`
		CreatedBy     string      `json:"createdBy,omitempty"`
		Default       *bool       `json:"default,omitempty"`
		Description   string      `json:"description,omitempty"`
		DisplayName   string      `json:"displayName,omitempty"`
		Id            string      `json:"id,omitempty"`
		LastUpdated   *time.Time  `json:"lastUpdated,omitempty"`
		LastUpdatedBy string      `json:"lastUpdatedBy,omitempty"`
		Name          string      `json:"name,omitempty"`
	}

	type User struct {
		Embedded              interface{}      `json:"_embedded,omitempty"`
		Links                 interface{}      `json:"_links,omitempty"`
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
	var got map[string]interface{}

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
