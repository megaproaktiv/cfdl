package cfdl_test

import (
	"github.com/megaproaktiv/cfdl"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"gotest.tools/assert"
)

// Mon Jan 2 15:04:05 MST 2006
const layoutAWS = "2006-01-02T15:04:05.000000-07:00"


func TestPopulateData(t *testing.T) {
	var countState = 0;
	// make and configure a mocked DeployInterface
	mockedDeployInterface := &cfdl.DeployInterfaceMock{
		CreateStackFunc: func(ctx context.Context, params *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
				panic("mock out the CreateStack method")
		},
		DeleteStackFunc: func(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
				panic("mock out the DeleteStack method")
		},
		DescribeStackEventsFunc: func(ctx context.Context, params *cloudformation.DescribeStackEventsInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
			var Events cloudformation.DescribeStackEventsOutput;
			var data []byte;
			var err error;
			if( countState == 0){
				data, err = ioutil.ReadFile("test/events1.json")
				countState++
			}else if( countState == 1){
				data, err = ioutil.ReadFile("test/events2.json")
				countState++
			}
			if err != nil {
				fmt.Println("File reading error", err)
			}
			json.Unmarshal(data, &Events);

			return &Events,nil;

		},
	}

	dataPre := map[string]cfdl.CloudFormationResource{
			"testcfn" : {
				LogicalResourceID: "testfncn",
				Type: "AWS::CloudFormation::Stack",
			},
			"MyTopic" : {
				LogicalResourceID: "MyTopic",
				Type: "AWS::SNS::Topic",
			},
			"NotMyTopic" : {
				LogicalResourceID: "NotMyTopic",
				Type: "AWS::SNS::Topic",
			},
	}

	// Timestamps from events1.json
	t1, _ := time.Parse(layoutAWS, "2020-11-06T10:55:46.074000+00:00");
	t2, _ := time.Parse(layoutAWS, "2020-11-06T10:55:49.190000+00:00");
	t3, _ := time.Parse(layoutAWS, "2020-11-06T10:55:49.187000+00:00")
	dataTarget1 := map[string]cfdl.CloudFormationResource{
		"testcfn" : {
			LogicalResourceID: "testfncn",
			PhysicalResourceID: "arn:aws:cloudformation:eu-central-1:123456789012:stack/testcfn/9f675870-201e-11eb-a9a9-06cc4e94edaa",
			Status: "CREATE_IN_PROGRESS",
			Type: "AWS::CloudFormation::Stack",
			Timestamp: t1,
			ResourceStatusReason: "User Initiated",
		},
		"MyTopic" : {
			LogicalResourceID: "MyTopic",
			PhysicalResourceID: "arn:aws:sns:eu-central-1:123456789012:my-topic-1604660145",
			Status: "CREATE_IN_PROGRESS",
			Type: "AWS::SNS::Topic",
			Timestamp: t2,
			ResourceStatusReason: "Resource creation Initiated",

		},
		"NotMyTopic" : {
			LogicalResourceID: "NotMyTopic",
			PhysicalResourceID: "arn:aws:sns:eu-central-1:123456789012:my-topic2-1604660145",
			Status: "CREATE_IN_PROGRESS",
			Type: "AWS::SNS::Topic",
			Timestamp: t3,
			ResourceStatusReason: "Resource creation Initiated",
		},
	}
	
	t1, _ = time.Parse(layoutAWS, "2020-11-06T10:56:00.644000+00:00");
	t2, _ = time.Parse(layoutAWS, "2020-11-06T10:55:59.605000+00:00");
	t3, _ = time.Parse(layoutAWS, "2020-11-06T10:55:59.693000+00:00")
	dataTarget2 := map[string]cfdl.CloudFormationResource{
			"testcfn" : {
				LogicalResourceID: "testfncn",
				PhysicalResourceID: "arn:aws:cloudformation:eu-central-1:012345678912:stack/testcfn/9f675870-201e-11eb-a9a9-06cc4e94edaa",
				Status: "CREATE_COMPLETE",
				Type: "AWS::CloudFormation::Stack",
				ResourceStatusReason: "User Initiated",
				Timestamp: t1,
			},
			"MyTopic" : {
				LogicalResourceID: "MyTopic",
				PhysicalResourceID: "arn:aws:sns:eu-central-1:012345678912:my-topic-1604660145",
				Status: "CREATE_COMPLETE",
				Type: "AWS::SNS::Topic",
				Timestamp: t2,
				ResourceStatusReason: "Resource creation Initiated",
			},
			"NotMyTopic" : {
				LogicalResourceID: "NotMyTopic",
				PhysicalResourceID: "arn:aws:sns:eu-central-1:012345678912:my-topic2-1604660145",
				Status:"CREATE_COMPLETE",
				Type: "AWS::SNS::Topic",
				Timestamp: t3,
				ResourceStatusReason: "Resource creation Initiated",
			},
	}

	data1 := cfdl.PopulateData(mockedDeployInterface, "TestStack", dataPre);
	assert.DeepEqual(t,dataTarget1, data1)

	data2 := cfdl.PopulateData(mockedDeployInterface, "TestStack", data1);
	assert.DeepEqual(t,dataTarget2, data2)
}

func TestIsStackCompleted(t *testing.T){
// Timestamps from events1.json
t1, _ := time.Parse(layoutAWS, "2020-11-06T10:55:46.074000+00:00");
t2, _ := time.Parse(layoutAWS, "2020-11-06T10:55:49.190000+00:00");
t3, _ := time.Parse(layoutAWS, "2020-11-06T10:55:49.187000+00:00")
dataTarget1 := map[string]cfdl.CloudFormationResource{
	"testcfn" : {
		LogicalResourceID: "testfncn",
		PhysicalResourceID: "",
		Status: "CREATE_IN_PROGRESS",
		Type: "AWS::CloudFormation::Stack",
		Timestamp: t1,
	},
	"MyTopic" : {
		LogicalResourceID: "MyTopic",
		Status: "CREATE_IN_PROGRESS",
		Type: "AWS::SNS::Topic",
		Timestamp: t2,
	},
	"NotMyTopic" : {
		LogicalResourceID: "NotMyTopic",
		Status: "CREATE_IN_PROGRESS",
		Type: "AWS::SNS::Topic",
		Timestamp: t3,
	},
}

t1, _ = time.Parse(layoutAWS, "2020-11-06T10:56:00.644000+00:00");
t2, _ = time.Parse(layoutAWS, "2020-11-06T10:55:59.605000+00:00");
t3, _ = time.Parse(layoutAWS, "2020-11-06T10:55:59.693000+00:00")
dataTarget2 := map[string]cfdl.CloudFormationResource{
		"testcfn" : {
			LogicalResourceID: "testfncn",
			PhysicalResourceID: "",
			Status: "CREATE_COMPLETE",
			Type: "AWS::CloudFormation::Stack",
			Timestamp: t1,
		},
		"MyTopic" : {
			LogicalResourceID: "MyTopic",
			Status: "CREATE_COMPLETE",
			Type: "AWS::SNS::Topic",
			Timestamp: t2,
		},
		"NotMyTopic" : {
			LogicalResourceID: "NotMyTopic",
			Status:"CREATE_COMPLETE",
			Type: "AWS::SNS::Topic",
			Timestamp: t3,
		},
}	
dataTarget3 := map[string]cfdl.CloudFormationResource{
		"testcfn" : {
			LogicalResourceID: "testfncn",
			PhysicalResourceID: "",
			Status: "CREATE_COMPLETE",
			Type: "AWS::CloudFormation::Stack",
			Timestamp: t1,
		},
		"MyTopic" : {
			LogicalResourceID: "MyTopic",
			Status: "UPDATE_COMPLETE",
			Type: "AWS::SNS::Topic",
			Timestamp: t2,
		},
		"NotMyTopic" : {
			LogicalResourceID: "NotMyTopic",
			Status:"CREATE_COMPLETE",
			Type: "AWS::SNS::Topic",
			Timestamp: t3,
		},
}	
	
dataTarget4 := map[string]cfdl.CloudFormationResource{
		"testcfn" : {
			LogicalResourceID: "testfncn",
			PhysicalResourceID: "",
			Status: "CREATE_IN_PROGRESS",
			Type: "AWS::CloudFormation::Stack",
			Timestamp: t1,
		},
		"MyTopic" : {
			LogicalResourceID: "MyTopic",
			Status: "UPDATE_COMPLETE",
			Type: "AWS::SNS::Topic",
			Timestamp: t2,
		},
		"NotMyTopic" : {
			LogicalResourceID: "NotMyTopic",
			Status:"CREATE_COMPLETE",
			Type: "AWS::SNS::Topic",
			Timestamp: t3,
		},
}	
	complete1 := cfdl.IsStackCompleted(dataTarget1);
	assert.Equal(t,false, complete1)
	
	complete2 := cfdl.IsStackCompleted(dataTarget2);
	assert.Equal(t,true, complete2)
	
	complete3 := cfdl.IsStackCompleted(dataTarget3);
	assert.Equal(t,true, complete3)

	assert.Equal(t,0,cfdl.CountCompleted(dataTarget1))
	assert.Equal(t,3,cfdl.CountCompleted(dataTarget2))
	assert.Equal(t,3,cfdl.CountCompleted(dataTarget3))
	assert.Equal(t,2,cfdl.CountCompleted(dataTarget4))
	
}