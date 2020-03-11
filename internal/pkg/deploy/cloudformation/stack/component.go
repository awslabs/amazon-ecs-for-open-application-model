package stack

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation/types"
	"github.com/gobuffalo/packd"
)

const (
	templatePath = "core.oam.dev/cf.yml"
)

// ComponentStackConfig is for providing all the values to set up an
// component instance stack and to interpret the outputs from it.
type ComponentStackConfig struct {
	*types.DeployComponentInput
	box packd.Box
}

// NewComponentStackConfig sets up a struct which can provide values to CloudFormation for
// spinning up a component instance.
func NewComponentStackConfig(input *types.DeployComponentInput, box packd.Box) *ComponentStackConfig {
	return &ComponentStackConfig{
		DeployComponentInput: input,
		box:                  box,
	}
}

// Template returns the component instance CloudFormation template.
func (e *ComponentStackConfig) Template() (string, error) {
	workloadTemplate, err := e.box.FindString(templatePath)
	if err != nil {
		return "", &ErrTemplateNotFound{templateLocation: templatePath, parentErr: err}
	}

	template, err := template.New("template").
		Funcs(templateFunctions).
		Funcs(sprig.FuncMap()).
		Parse(workloadTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := template.Execute(&buf, e.DeployComponentInput); err != nil {
		return "", err
	}

	return string(buf.Bytes()), nil
}

// Parameters returns the parameters to be passed into a component instance CloudFormation template.
func (e *ComponentStackConfig) Parameters() []*cloudformation.Parameter {
	return []*cloudformation.Parameter{}
}

// Tags returns the tags that should be applied to the component instance CloudFormation stack.
func (e *ComponentStackConfig) Tags() []*cloudformation.Tag {
	return []*cloudformation.Tag{
		{
			Key:   aws.String(ComponentTagKey),
			Value: aws.String(e.ComponentConfiguration.InstanceName),
		},
		{
			Key:   aws.String(AppTagKey),
			Value: aws.String(e.ApplicationConfiguration.Name),
		},
		{
			Key:   aws.String(EnvTagKey),
			Value: aws.String(EnvTagValue),
		},
	}
}

// StackName returns the name of the CloudFormation stack (hard-coded).
func (e *ComponentStackConfig) StackName() string {
	const maxLen = 128
	stackName := fmt.Sprintf("oam-ecs-%s-%s", e.ApplicationConfiguration.Name, e.ComponentConfiguration.InstanceName)
	if len(stackName) > maxLen {
		return stackName[len(stackName)-maxLen:]
	}
	return stackName
}

// ToComponent inspects a component instance cloudformation stack and constructs a component instance
// struct out of it
func (e *ComponentStackConfig) ToComponent(stack *cloudformation.Stack) (*types.Component, error) {
	outputs := map[string]string{}

	for _, output := range stack.Outputs {
		key := *output.OutputKey
		value := *output.OutputValue
		outputs[key] = value
	}

	createdComponent := types.Component{
		StackName:    e.StackName(),
		StackOutputs: outputs,
	}

	return &createdComponent, nil
}
