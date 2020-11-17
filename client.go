package cfdl

import (
	config "github.com/aws/aws-sdk-go-v2/config"
	cfnservice "github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

// Client get a CloudFormation service client
func Client() (*cfnservice.Client){
	cfg, err := config.LoadDefaultConfig(config.WithRegion("eu-central-1"))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}	
	client := cfnservice.NewFromConfig(cfg);
	return client
}