#!/bin/bash

export PATH="$PATH:$PWD/bin/local"
PS1="$ "

. ../demo-magic/demo-magic.sh

clear

# Clean up previous runs

rm -rf examples/oam-ecs-dry-run-results

# Look and feel

TYPE_SPEED=20
DEMO_COMMENT_COLOR=$CYAN
NO_WAIT=false

# Start the demo

PROMPT_TIMEOUT=0
p "# Welcome to oam-ecs!"
PROMPT_TIMEOUT=1

NO_WAIT=true
p "# The oam-ecs CLI is a proof-of-concept that partially implements the Open Application Model (OAM)"
p "# specification, version v1alpha1. oam-ecs takes OAM definitions as input, translates them into AWS"
p "# CloudFormation templates, and deploys them as Amazon ECS services running on AWS Fargate.\n"
NO_WAIT=false

p "# Let's walk through an example!"

pe "cd examples/"

pe "ls -1"

p "# oam-ecs works with two types of files: component configuration and application configuration."

p "# Components are things like a backend API service, a frontend web application, or a database."

pe "ls -1 *-component.yaml"

NO_WAIT=true
p "# Here I have two components: a frontend 'server' workload that has an HTTP endpoint, and a "
p "# backend 'worker' workload that processes data."
NO_WAIT=false

pe "head -n 15 server-component.yaml"

NO_WAIT=true
p "# Component configuration like the example above contains information from the developer about "
p "# requirements for running their application code, like the image and resource requirements."
NO_WAIT=false

pe "cat example-app.yaml"

NO_WAIT=true
p "# Application configuration like the example above deploys new instances of the developer's "
p "# components. It contains information from the operator that is specific to each instance of the "
p "# component, like environment variable values and scale.\n"

p "# Note that neither the component config nor the application config specified *any* infrastructure!"
p "# The OAM format is platform-agnostic, so infrastructure operators decide what infrastructure this "
p "# configuration should translate into. For example, oam-ecs runs OAM workloads with ECS on Fargate.\n"
NO_WAIT=false

p "# So, let's deploy this application to ECS and Fargate with oam-ecs!"

p "# I already created an oam-ecs environment, which contains shared infrastructure like the VPC."

pe "oam-ecs env show"

NO_WAIT=true
p "# I can do a dry-run and inspect the CloudFormation templates that my OAM configuration produces, "
p "# before deploying them."
NO_WAIT=false

pe "oam-ecs app deploy --dry-run -f example-app.yaml -f worker-component.yaml -f server-component.yaml"

pe "ls -1 oam-ecs-dry-run-results"

pe "less oam-ecs-dry-run-results/oam-ecs-example-app-example-server-template.yaml"

p "# Now I will deploy the infrastructure and my application containers into my AWS account."

pe "oam-ecs app deploy -f example-app.yaml -f worker-component.yaml -f server-component.yaml"

NO_WAIT=true
p "# I can see details about the deployed resources, including the endpoint that was created for my "
p "# server workload.\n"
NO_WAIT=false

PROMPT_TIMEOUT=0
p "# Enjoy deploying OAM applications to ECS on Fargate with oam-ecs!"
