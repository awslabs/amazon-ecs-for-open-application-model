// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cloudformation provides functionality to deploy oam-ecs resources with AWS CloudFormation.
package cloudformation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/awslabs/amazon-ecs-for-open-application-model/templates"
	"github.com/gobuffalo/packd"
)

type stackConfiguration interface {
	StackName() string
	Template() (string, error)
	Parameters() []*cloudformation.Parameter
	Tags() []*cloudformation.Tag
}

// regionalClientProvider lets us make cross region describe calls
// in one CloudFormation struct. We can dynamically generate clients
// configured for a specific region.
type regionalClientProvider interface {
	Client(string) cloudformationiface.CloudFormationAPI
}

type cfClientBuilder struct {
	session *session.Session
}

func (cf cfClientBuilder) Client(region string) cloudformationiface.CloudFormationAPI {
	return cloudformation.New(cf.session, &aws.Config{Region: aws.String(region)})
}

// CloudFormation wraps the CloudFormationAPI interface
type CloudFormation struct {
	regionalClientProvider regionalClientProvider
	waiters                []request.WaiterOption
	client                 cloudformationiface.CloudFormationAPI
	box                    packd.Box
}

// New returns a configured CloudFormation client.
func New(sess *session.Session) CloudFormation {
	cb := cfClientBuilder{
		session: sess,
	}

	waiterOptions := []request.WaiterOption{
		// Poll for CloudFormation updates every 3 seconds.
		request.WithWaiterDelay(request.ConstantWaiterDelay(3 * time.Second)),
		// Wait for at most 90 mins for any CloudFormation action.
		request.WithWaiterMaxAttempts(1800),
	}

	return CloudFormation{
		regionalClientProvider: cb,
		client:                 cb.Client(*sess.Config.Region),
		box:                    templates.Box(),
		waiters:                waiterOptions,
	}
}

func (cf CloudFormation) waitForStackCreation(stackConfig stackConfiguration) (*cloudformation.Stack, error) {
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackConfig.StackName()),
	}

	if err := cf.client.WaitUntilStackCreateCompleteWithContext(context.Background(), describeStackInput, cf.waiters...); err != nil {
		return nil, fmt.Errorf("failed to create stack %s: %w", stackConfig.StackName(), err)
	}

	return cf.describeStack(describeStackInput)
}

func (cf CloudFormation) waitForStackUpdate(stackConfig stackConfiguration) (*cloudformation.Stack, error) {
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackConfig.StackName()),
	}

	if err := cf.client.WaitUntilStackUpdateCompleteWithContext(context.Background(), describeStackInput, cf.waiters...); err != nil {
		return nil, fmt.Errorf("failed to update stack %s: %w", stackConfig.StackName(), err)
	}

	return cf.describeStack(describeStackInput)
}

func (cf CloudFormation) describe(stackConfig stackConfiguration) (*cloudformation.Stack, error) {
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackConfig.StackName()),
	}

	return cf.describeStack(describeStackInput)
}

func (cf CloudFormation) describeStack(describeStackInput *cloudformation.DescribeStacksInput) (*cloudformation.Stack, error) {
	return cf.describeStackWithClient(describeStackInput, cf.client)
}

// describeStackWithClient let's us use a preconfigured client to make calls to CloudFormation.
// This is useful when we need to make cross-region calls.
func (cf CloudFormation) describeStackWithClient(describeStackInput *cloudformation.DescribeStacksInput,
	client cloudformationiface.CloudFormationAPI) (*cloudformation.Stack, error) {
	describeStackOutput, err := client.DescribeStacks(describeStackInput)
	if err != nil {
		if stackDoesNotExist(err) {
			return nil, &ErrStackNotFound{stackName: *describeStackInput.StackName}
		}
		return nil, err
	}

	if len(describeStackOutput.Stacks) == 0 {
		return nil, &ErrStackNotFound{stackName: *describeStackInput.StackName}
	}

	return describeStackOutput.Stacks[0], nil
}

// create will only spin up a stack if none exists or if a stack exists
// but requires cleanup (meaning we failed to created it before). With
// stacks that are failed to be created, you have to delete them if you
// want to update them. In this case, we'll delete the failed stack
// and then try creating it again.
// If a stack already exists in another state, we'll return an ErrStackAlreadyExists
// error.
func (cf CloudFormation) create(stackConfig stackConfiguration) (bool, error) {
	describeStackInput := &cloudformation.DescribeStacksInput{StackName: aws.String(stackConfig.StackName())}
	existingStack, err := cf.describeStack(describeStackInput)
	// Create the stack if it doesn't already exists.
	if err != nil {

		var stackNotFound *ErrStackNotFound
		if !errors.As(err, &stackNotFound) {
			return false, err
		}
		// If there's no existing stack, we can go ahead and create it.
		return cf.deploy(stackConfig, cloudformation.ChangeSetTypeCreate)
	}

	// If the stack exists, but failed to create, we'll clean it up and
	// then re-create it.
	if StackStatus(*existingStack.StackStatus).RequiresCleanup() {
		if err := cf.delete(stackConfig.StackName()); err != nil {
			return false, fmt.Errorf("cleaning up a previous failed stack: %w", err)
		}
		return cf.deploy(stackConfig, cloudformation.ChangeSetTypeCreate)
	}

	if StackStatus(*existingStack.StackStatus).InProgress() {
		return false, &ErrStackUpdateInProgress{
			stackName:   stackConfig.StackName(),
			stackStatus: aws.StringValue(existingStack.StackStatus),
		}
	}

	// If the stack exists and has been successfully created - return
	// a ErrStackAlreadyExists error.
	return false, &ErrStackAlreadyExists{
		stackName: stackConfig.StackName(),
		parentErr: fmt.Errorf("with status: %s", *existingStack.StackStatus),
	}
}

// update will update a given stack, so long as the stack already exists
// and it isn't already deploying something. If there's already some action
// happening on this stack, we'll return ErrStackUpdateInProgress.
func (cf CloudFormation) update(stackConfig stackConfiguration) (bool, error) {
	describeStackInput := &cloudformation.DescribeStacksInput{StackName: aws.String(stackConfig.StackName())}
	existingStack, err := cf.describeStack(describeStackInput)
	// If we can't find the stack to update, return an error.
	if err != nil {
		return false, err
	}

	// If the stack exists but is in progress, return an error.
	if StackStatus(*existingStack.StackStatus).InProgress() {
		return false, &ErrStackUpdateInProgress{
			stackName:   stackConfig.StackName(),
			stackStatus: aws.StringValue(existingStack.StackStatus),
		}
	}

	return cf.deploy(stackConfig, cloudformation.ChangeSetTypeUpdate)
}

func (cf CloudFormation) delete(stackName string) error {
	if _, err := cf.client.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	}); err != nil {
		return fmt.Errorf("deleting stack %s: %w", stackName, err)
	}

	return cf.client.WaitUntilStackDeleteCompleteWithContext(context.Background(),
		&cloudformation.DescribeStacksInput{StackName: aws.String(stackName)},
		cf.waiters...)
}

func (cf CloudFormation) deploy(stackConfig stackConfiguration, createOrUpdate string) (bool, error) {
	template, err := stackConfig.Template()
	if err != nil {
		return false, fmt.Errorf("template creation: %w", err)
	}

	in, err := createChangeSetInput(stackConfig.StackName(),
		template,
		withChangeSetType(createOrUpdate),
		withTags(stackConfig.Tags()),
		withParameters(stackConfig.Parameters()))

	if err != nil {
		return false, err
	}

	return cf.deployChangeSet(in)
}

// Returns true if deployment is in-progress, false if deployment failed or was not needed
func (cf CloudFormation) deployChangeSet(in *cloudformation.CreateChangeSetInput) (bool, error) {
	set, err := cf.createChangeSet(in)
	if err != nil {
		return false, err
	}
	if err := set.waitForCreation(); err != nil {
		// NOTE: If WaitUntilChangeSetCreateComplete returns an error it's possible that there
		// are simply no changes between the previous and proposed Stack ChangeSets. We make a call to
		// DescribeChangeSet to see if that is indeed the case and handle it gracefully.
		if err := set.describe(); err != nil {
			return false, fmt.Errorf("describing failed change set: %w", err)
		}

		// The change set was empty - so we clean it up and don't return an error.
		// We have to clean up the changeSet because there's a limit on the number
		// of failed changesets a customer can have on a particular stack.
		if len(set.changes) == 0 {
			set.delete()
			return false, nil
		}

		return false, err
	}
	if err := set.execute(); err != nil {
		return false, err
	}
	return true, nil
}

func (cf CloudFormation) createChangeSet(in *cloudformation.CreateChangeSetInput) (*changeSet, error) {
	out, err := cf.client.CreateChangeSet(in)
	if err != nil {
		return nil, fmt.Errorf("failed to create changeSet for stack %s: %w", *in.StackName, err)
	}
	return &changeSet{
		name:    aws.StringValue(out.Id),
		stackID: aws.StringValue(out.StackId),
		c:       cf.client,
		waiters: cf.waiters,
	}, nil
}

// stackDoesNotExist returns true if the underlying error is a stack doesn't exist.
func stackDoesNotExist(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case "ValidationError":
			// A ValidationError occurs if we describe a stack which doesn't exist.
			if strings.Contains(aerr.Message(), "does not exist") {
				return true
			}
		}
	}
	return false
}
