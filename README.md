# cfdl AWS CloudFormation Deploy library

Deploy CloudFormation as go programm.

## Usage

1) create a CloudFormation client

```go
client := cfdl.Client()
```    

2) create a goformation template

```go
import 	"github.com/awslabs/goformation/v4/cloudformation"
...	
template :=	template := cloudformation.NewTemplate()
...
```

3) Populate template

```go
import 	"github.com/awslabs/goformation/v4/cloudformation/sns"
...
	template.Resources["GoFormationCompareTopic"] = &sns.Topic{
		TopicName: "CfdCdkCompareTopic" + strconv.FormatInt(time.Now().Unix(), 10),
		};
```

4) Deploy stack

```go
	cfdl.CreateStack(client,stackname, template)
```