package goflat

import (
	"fmt"
	"reflect"
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
