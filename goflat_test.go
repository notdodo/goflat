package goflat

import (
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		test     string
		expected string
	}{
		{
			test:     `{"a":"3","c":4,"b":{"d":"5","e":6}}`,
			expected: `{"a":"3","b.d":"5","b.e":6,"c":4}`,
		},
		{
			test:     `{"a": "3", "b": {"c":true}}`,
			expected: `{"a":"3","b.c":true}`,
		},
		{
			test:     `[{"a": "3"}, {"b": "3", "C": [{"c": 10}, {"d": 11}]}]`,
			expected: `[{"a":"3"},{"C.c.0":10,"C.d.1":11,"b":"3"}]`,
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
			expected: `[{"InlinePolicies.PolicyName.0":"policy-s3-operator","InlinePolicies.Statement.Action.0.0.0":"s3:ListAllMyBuckets","InlinePolicies.Statement.Action.0.1.0":"s3:ListBucket","InlinePolicies.Statement.Action.0.2.0":"s3:PutObject","InlinePolicies.Statement.Action.1.1.0":"s3:GetBucketLocation","InlinePolicies.Statement.Action.1.2.0":"s3:GetObject","InlinePolicies.Statement.Action.2.2.0":"s3:AbortMultipartUpload","InlinePolicies.Statement.Action.3.2.0":"s3:ListMultipartUploadParts","InlinePolicies.Statement.Action.4.2.0":"s3:ListBucketMultipartUploads","InlinePolicies.Statement.Effect.0.0":"Allow","InlinePolicies.Statement.Effect.1.0":"Allow","InlinePolicies.Statement.Effect.2.0":"Allow","InlinePolicies.Statement.Resource.0.0.0":"arn:aws:s3:::*","InlinePolicies.Statement.Resource.0.1.0":"arn:aws:s3:::personal-s3-bucket/*","InlinePolicies.Statement.Resource.0.2.0":"arn:aws:s3:::personal-s3-bucket/*","UserId":"AIDARRRRRRRRRRRR","UserName":"s3-operator"}]`},
	}

	for _, test := range tests {
		got, err := Flat(test.test, "", ".")
		if err != nil {
			t.Errorf("[X] Test failed with error: %v", err)
			continue
		}
		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("mismatch, got: %v wanted: %v", got, test.expected)
		}
	}
}
