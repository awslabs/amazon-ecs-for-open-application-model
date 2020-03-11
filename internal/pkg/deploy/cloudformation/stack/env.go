package stack

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation/types"
	"github.com/gobuffalo/packd"
)

// EnvStackConfig is for providing all the values to set up an
// environment stack and to interpret the outputs from it.
type EnvStackConfig struct {
	*types.CreateEnvironmentInput
	box packd.Box
}

const (
	// EnvTemplatePath is the path where the cloudformation for the environment is written.
	EnvTemplatePath = "environment/cf.yml"
)

// NewEnvStackConfig sets up a struct which can provide values to CloudFormation for
// spinning up an environment.
func NewEnvStackConfig(input *types.CreateEnvironmentInput, box packd.Box) *EnvStackConfig {
	return &EnvStackConfig{
		CreateEnvironmentInput: input,
		box:                    box,
	}
}

// Template returns the environment CloudFormation template.
func (e *EnvStackConfig) Template() (string, error) {
	environmentTemplate, err := e.box.FindString(EnvTemplatePath)
	if err != nil {
		return "", &ErrTemplateNotFound{templateLocation: EnvTemplatePath, parentErr: err}
	}

	return environmentTemplate, nil
}

// Parameters returns the parameters to be passed into a environment CloudFormation template.
func (e *EnvStackConfig) Parameters() []*cloudformation.Parameter {
	return []*cloudformation.Parameter{}
}

// Tags returns the tags that should be applied to the environment CloudFormation stack.
func (e *EnvStackConfig) Tags() []*cloudformation.Tag {
	return []*cloudformation.Tag{
		{
			Key:   aws.String(EnvTagKey),
			Value: aws.String(EnvTagValue),
		},
	}
}

// StackName returns the name of the CloudFormation stack (hard-coded).
func (e *EnvStackConfig) StackName() string {
	return fmt.Sprintf("%s-%s", EnvTagKey, EnvTagValue)
}

// ToEnv inspects an environment cloudformation stack and constructs an environment
// struct out of it
func (e *EnvStackConfig) ToEnv(stack *cloudformation.Stack) (*types.Environment, error) {
	outputs := map[string]string{}

	for _, output := range stack.Outputs {
		key := *output.OutputKey
		value := *output.OutputValue
		outputs[key] = value
	}

	createdEnv := types.Environment{
		StackName:    e.StackName(),
		StackOutputs: outputs,
	}

	return &createdEnv, nil
}
