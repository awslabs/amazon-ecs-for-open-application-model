package types

// CreateEnvironmentInput holds the fields required to deploy an environment.
type CreateEnvironmentInput struct {
}

// Environment represents the configuration of a particular environment
type Environment struct {
	StackName    string
	StackOutputs map[string]string
}

// CreateEnvironmentResponse holds the created environment on successful deployment.
// Otherwise, the environment is set to nil and a descriptive error is returned.
type CreateEnvironmentResponse struct {
	Env *Environment
	Err error
}
