package cfdl

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	cfn "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation"
	tm "github.com/buger/goterm"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

// StatusCreateComplete CloudFormation Status
const StatusCreateComplete="CREATE_COMPLETE"
// StatusCreateInProgress CloudFormation Status
const StatusCreateInProgress = "CREATE_IN_PROGRESS"
// StatusDeleteComplete CloudFormation Status
const StatusDeleteComplete = "DELETE_COMPLETE"
// StatusUpdateComplete text for update
const StatusUpdateComplete = "UPDATE_COMPLETE"

const (
	// ColorDefault default color
	ColorDefault = "\x1b[39m"
	// ColorRed red for screen
	ColorRed   = "\x1b[91m"
	// ColorGreen green for screen
	ColorGreen = "\x1b[32m"
	// ColorBlue blue for screen
	ColorBlue  = "\x1b[94m"
	// ColorGray for screen
	ColorGray  = "\x1b[90m"
)

// CloudFormationResource holder for status
type CloudFormationResource struct {
	LogicalResourceID string
	PhysicalResourceID string
	Status string
	Type string
	Timestamp time.Time
	ResourceStatusReason string
}

//go:generate moq -out deploy_moq_test.go . DeployInterface

// DeployInterface all deployment functions
type DeployInterface interface {
	CreateStack(ctx context.Context, params *cfn.CreateStackInput, optFns ...func(*cfn.Options)) (*cfn.CreateStackOutput, error)
	DescribeStackEvents(ctx context.Context, params *cfn.DescribeStackEventsInput, optFns ...func(*cfn.Options)) (*cfn.DescribeStackEventsOutput, error)
	DeleteStack(ctx context.Context, params *cfn.DeleteStackInput, optFns ...func(*cfn.Options)) (*cfn.DeleteStackOutput, error)
	UpdateStack(ctx context.Context, params *cfn.UpdateStackInput, optFns ...func(*cfn.Options)) (*cfn.UpdateStackOutput, error) 
	CreateChangeSet(ctx context.Context, params *cfn.CreateChangeSetInput, optFns ...func(*cfn.Options)) (*cfn.CreateChangeSetOutput, error)
	ExecuteChangeSet(ctx context.Context, params *cfn.ExecuteChangeSetInput, optF ...func(*cfn.Options)) (*cfn.ExecuteChangeSetOutput, error)
}

// CreateStack first time stack creation
// client - the aws cloudformation client
// name - name if the Cloudformaiton template
// template - a goformation template
func CreateStack(client DeployInterface,name string, template *cloudformation.Template){
	DumpTemplate(template)
	stack, _ := template.YAML()
	templateBody := string(stack)

	params := &cfn.CreateStackInput{
		StackName: &name,
		TemplateBody: &templateBody,		
	}
	slog.Info("CreateStack: ",name)
	response, err := client.CreateStack(context.TODO(),params)
	if err != nil {
		slog.Error("CreateStack ",err.Error())
		panic(err)
	}
	slog.Debug("Response ",response)
}

// UpdateStack first time stack creation
// client - the aws cloudformation client
// name - name if the Cloudformaiton template
// template - a goformation template
func UpdateStack(client DeployInterface,name string, template *cloudformation.Template){
	DumpTemplate(template)
	stack, _ := template.YAML()
	templateBody := string(stack)

	params := &cfn.UpdateStackInput{
		StackName: &name,
		TemplateBody: &templateBody,		
	}
	slog.Info("UpdateStack: ",name)
	response, err := client.UpdateStack(context.TODO(),params)
	if err != nil {
		slog.Error("CreateStack ",err.Error())
		panic(err)
	}
	slog.Debug("Response ",response)
}


// CreateChangeSet start an Change Set cycle
func CreateChangeSet(client DeployInterface,name string, template *cloudformation.Template){
	DumpTemplate(template)
	stack, _ := template.YAML()
	templateBody := string(stack)
    uuidWithHyphen := uuid.New()
	changeSetName := name+"-"+uuidWithHyphen.String();

	//template.Resources["changesetid"] = cloudformation.NewTemplate().Metadata
	params := &cfn.CreateChangeSetInput{
		StackName: &name,
		TemplateBody: &templateBody,	
		ChangeSetName: &changeSetName,
	}
	slog.Info("UpdateStack: ",name)
	response, err := client.CreateChangeSet(context.TODO(),params)
	if err != nil {
		slog.Error("CreateStack ",err.Error())
		panic(err)
	}
	slog.Debug("Response ",response)
}

// func ExecuteChangeSet(client DeployInterface,name string){

// }


// DeleteStack first time stack creation
func DeleteStack(client DeployInterface,name string){

	params := &cfn.DeleteStackInput{
		StackName: &name,
	}

	client.DeleteStack(context.TODO(),params)
}


// ShowStatus status of stack
func ShowStatus(client DeployInterface, name string, template *cloudformation.Template, endState string){
	
	// Prepopulate
	
    data := map[string]CloudFormationResource{}
	i := 1
	for k, v := range template.Resources {
		i = i+1
		item := &CloudFormationResource{
			LogicalResourceID: k,
			PhysicalResourceID: "",
			Status: "-",
			Type: v.AWSCloudFormationType(),
		}
		data[k] = *item;
	}
	
	// Draw
	table := simpletable.New()
	errorTable := simpletable.New();

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "ID"},
			{Align: simpletable.AlignLeft, Text: "Status"},
			{Align: simpletable.AlignLeft, Text: "Type"},
			{Align: simpletable.AlignLeft, Text: "PhysicalResourceID"},
		},
		
	}
	table.SetStyle(simpletable.StyleCompactLite)
	
	errorTable.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "ID"},
			{Align: simpletable.AlignLeft, Text: "Status"},
			{Align: simpletable.AlignLeft, Text: "Status Reason"},
		},
		
	}
	errorTable.SetStyle(simpletable.StyleCompactLite)

	first := true
	firstError := true
	for !IsStackCompleted(data){
		tm.Clear()
		tm.MoveCursor(1,1)
		data = PopulateData(client, name, data);
		i = 0;
		j := 0;
		var statustext string

		// Sort
		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			v := data[k]
			id := data[k].LogicalResourceID
			if( v.Status == StatusCreateComplete){
				statustext = green(StatusCreateComplete)
			}else if v.Status == StatusDeleteComplete {
				statustext = red(StatusDeleteComplete)
			} else{		
				statustext = gray(v.Status)
			}

			r := []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: id},
				{Align: simpletable.AlignLeft, Text: statustext},
				{Align: simpletable.AlignLeft, Text: v.Type},
				{Align: simpletable.AlignLeft, Text: v.PhysicalResourceID},
			}
			
			if len(v.ResourceStatusReason) > 0  {
				re := []*simpletable.Cell{
					{Align: simpletable.AlignLeft, Text: id},
					{Align: simpletable.AlignLeft, Text: statustext},
					{Align: simpletable.AlignLeft, Text: v.ResourceStatusReason},
				}
				if !firstError {
					errorTable.Body.Cells[j] = re
				}else{
					firstError = false
					errorTable.Body.Cells = append(errorTable.Body.Cells,re)
				}
				j = j+1;
			}
			if !first {
				table.Body.Cells[i]=r
				}else{
					table.Body.Cells = append(table.Body.Cells, r)
			}	

			i = i+1;
		}
		first = false
		tm.Println(table.String())
		tm.Println()

		max := len(keys)
		current := CountCompleted(data)

		tm.Print("[")
		for i := 0; i < max; i++ {
			if i < current {
				tm.Print("X")
			}else {
				tm.Print("-")
			}
		}
		tm.Println("]")
		tm.Println()
		tm.Println(errorTable.String())
		tm.Flush()
		time.Sleep(1 * time.Second) 
	}
	
	
}


// PopulateData update status from describe call
func PopulateData(client DeployInterface, name string,data map[string]CloudFormationResource)( map[string]CloudFormationResource){
	params := &cfn.DescribeStackEventsInput{
		StackName: &name,
	}
	output, error := client.DescribeStackEvents(context.TODO(), params)
	if( error != nil){
		msg  := error.Error()
		if strings.Contains(msg, "does not exist"){
			fmt.Println("Stack <",name,"> does not exist");
			os.Exit(0);
		}

		panic(error)
	}

	// Update Status and Timestamp if newer
	for i := 0; i < len(output.StackEvents); i++ {
		
		event := output.StackEvents[i];		
		item := data[*event.LogicalResourceId]

		if( event.Timestamp.After(item.Timestamp) ){
			item.Status = string(event.ResourceStatus);
			item.Timestamp = *event.Timestamp;
			item.PhysicalResourceID = *event.PhysicalResourceId
			item.Type = *event.ResourceType
			if event.ResourceStatusReason != nil {
				item.ResourceStatusReason = *event.ResourceStatusReason
			}
			data[*event.LogicalResourceId] = item;

			
		}
		
	}
	return data;

}

func isComplete(status string) bool {
	return strings.HasSuffix(status , "_COMPLETE")
}

// IsStackCompleted check for everything "completed"
func IsStackCompleted(data map[string]CloudFormationResource) bool {
	for _, value := range data {
		if(!isComplete(value.Status)){
			return false
		}
	}
	return true;
}


// CountCompleted how many completed
func CountCompleted(data map[string]CloudFormationResource) int {
	var count int
	count = 0
	for _, value := range data {
		if isComplete(value.Status){
			count++
		}
	}
	return count
}

func red(s string) string {
	return fmt.Sprintf("%s%s%s", ColorRed, s, ColorDefault)
}

func green(s string) string {
	return fmt.Sprintf("%s%s%s", ColorGreen, s, ColorDefault)
}

func blue(s string) string {
	return fmt.Sprintf("%s%s%s", ColorBlue, s, ColorDefault)
}

func gray(s string) string {
	return fmt.Sprintf("%s%s%s", ColorGray, s, ColorDefault)
}

// DumpTemplate for debugging
func DumpTemplate(template *cloudformation.Template){
	y,_ := template.YAML()
	path := "dump"
	fullPath := "dump/template.yml"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
	f, err := os.Create(fullPath)
    if err != nil {
        fmt.Println(err)
        return
    }
    _, err = f.WriteString(string(y))
    if err != nil {
        fmt.Println(err)
        f.Close()
        return
    }
    err = f.Close()
    if err != nil {
        fmt.Println(err)
        return
	}
	fmt.Println("Template dumped in :","dump/template.yaml")
}