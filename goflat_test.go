package goflat

import (
	"fmt"
	"testing"

	oj "github.com/ohler55/ojg/oj"
	"github.com/r3labs/diff"
)

func TestFlattenOne(t *testing.T) {
	tests := []struct {
		test     string
		expected map[string]interface{}
	}{
		{
			test: `{"a":"3","c":4,"b":{"d":"5","e":6}}`,
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
			test: `[{"a": "3"}, {"b": "3", "C": [{"c": 10}, {"d": 11}]}]`,
			expected: map[string]interface{}{
				"a":     "3",
				"b":     "3",
				"C.c.0": int64(10),
				"C.d.1": int64(11),
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
				"UserId":                                  "AIDARRRRRRRRRRRR",
				"InlinePolicies.PolicyName.0":             "policy-s3-operator",
				"InlinePolicies.Statement.Action.4.2.0":   "s3:ListBucketMultipartUploads",
				"InlinePolicies.Statement.Effect.1.0":     "Allow",
				"InlinePolicies.Statement.Resource.0.1.0": "arn:aws:s3:::personal-s3-bucket/*",
			},
		},
	}

	for _, test := range tests {
		got, err := Flat(test.test, "", ".")
		if err != nil {
			t.Errorf("[X] Test failed with error: %v", err)
			continue
		}

		gotMap, err := oj.ParseString(got)
		if err != nil {
			t.Errorf("[X] Test failed with error: %v", err)
			continue
		}

		switch gotMapType := gotMap.(type) {
		case []interface{}: // [{"a": 1}]
			for testKey, testValue := range test.expected {
				testOk := false
				for _, gotMapTypeObject := range gotMapType {
					if val, ok := gotMapTypeObject.(map[string]interface{})[testKey]; ok {
						if val == testValue {
							testOk = true
						}
					}
				}
				if !testOk {
					t.Errorf("mismatch, got: %v wanted: %v:%v", gotMapType, testKey, testValue)
					continue
				}
			}
		case map[string]interface{}: // {"a": 1}
			for testKey, testValue := range test.expected {
				if val, ok := gotMapType[testKey]; ok {
					if val != testValue {
						t.Errorf("mismatch, got: %v wanted: %v", val, testValue)
						continue
					}
				} else {
					t.Errorf("key mismatch, got: %v wanted: %v", gotMapType, testKey)
					continue
				}
			}
		default:
			t.Errorf("type mismatch")
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
		Name string
		ID   int64
		Type TypeStr
	}{
		Name: "test",
		ID:   int64(54),
		Type: typeStr,
	}

	expectedMap := map[string]interface{}{
		prefix + "ID":                        int64(54),
		prefix + "Name":                      "test",
		prefix + "Type" + separator + "Name": "testflat",
	}

	a, _ := FlatStruct(testStruct, prefix, separator)
	diffs, _ := diff.Diff(a, expectedMap)

	if len(diffs) > 0 {
		fmt.Println(diffs)
		t.Errorf("map mismatch:\ngot: %v\nwanted: %v", a, expectedMap)
	}
}
