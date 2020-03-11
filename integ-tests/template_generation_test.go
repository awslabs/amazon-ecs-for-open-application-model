// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package template_generation_test

import (
	"fmt"
	"io/ioutil"

	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/cli"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

func MatchCloudFormationTemplate(expected interface{}) types.GomegaMatcher {
	return &CloudFormationTemplateMatcher{
		yamlMatcher: &matchers.MatchYAMLMatcher{},
		expected:    expected,
	}
}

type CloudFormationTemplateMatcher struct {
	yamlMatcher *matchers.MatchYAMLMatcher
	expected    interface{}
}

func (matcher *CloudFormationTemplateMatcher) Match(actual interface{}) (bool, error) {
	actualFileContents, err := ioutil.ReadFile(actual.(string))
	if err != nil {
		return false, err
	}

	expectedFileContents, err := ioutil.ReadFile(matcher.expected.(string))
	if err != nil {
		return false, err
	}

	matcher.yamlMatcher.YAMLToMatch = expectedFileContents
	return matcher.yamlMatcher.Match(actualFileContents)
}

func (matcher *CloudFormationTemplateMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Generated template %s does not match expected template %s", actual, matcher.expected)
}

func (matcher *CloudFormationTemplateMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected generated template %s to not match expected template %s", actual, matcher.expected)
}

var _ = Describe("generate dry-run CloudFormation templates", func() {

	var (
		applyOpts *cli.ApplyOpts
	)

	Context("Schematic validation", func() {
		BeforeEach(func() {
			applyOpts = cli.NewApplyOpts()
			applyOpts.DryRun = true
		})

		It("invalid API version should return an error", func() {
			applyOpts.OamFiles = []string{
				"schematics/wrong-api-version.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(MatchError(HavePrefix("no kind \"ComponentSchematic\" is registered for version \"core.oam.dev/hello\"")))
		})

		It("invalid kind should return an error", func() {
			applyOpts.OamFiles = []string{
				"schematics/wrong-kind.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(MatchError(HavePrefix("no kind \"HelloWorld\" is registered for version \"core.oam.dev/v1alpha1\"")))
		})

		It("invalid workload type should return an error", func() {
			applyOpts.OamFiles = []string{
				"schematics/wrong-workload-type.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(MatchError(HavePrefix("Workload type is core.oam.dev/v1alpha1.HelloWorld, only core.oam.dev/v1alpha1.Worker and core.oam.dev/v1alpha1.Server are supported")))
		})

		It("extended workload types are not supported", func() {
			applyOpts.OamFiles = []string{
				"schematics/extended-workload-type.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(MatchError(HavePrefix("Workload type is ecs.amazonaws.com/v1.ECSService, only core.oam.dev/v1alpha1.Worker and core.oam.dev/v1alpha1.Server are supported")))
		})

		It("application scopes are not supported", func() {
			applyOpts.OamFiles = []string{
				"schematics/application-scope.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(MatchError(HavePrefix("Object type core.oam.dev/v1alpha1, Kind=ApplicationScope is not supported")))
		})

		It("custom traits are not supported", func() {
			applyOpts.OamFiles = []string{
				"schematics/trait.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(MatchError(HavePrefix("Object type core.oam.dev/v1alpha1, Kind=Trait is not supported")))
		})

		It("multiple application configurations should return an error", func() {
			applyOpts.OamFiles = []string{
				"schematics/worker.yaml",
				"schematics/webserver.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(MatchError(HavePrefix("Multiple application configuration files found, only one is allowed per application")))
		})

		It("missing application configuration should return an error", func() {
			applyOpts.OamFiles = []string{
				"schematics/nginx.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(MatchError(HavePrefix("Application configuration is required")))
		})

		It("missing component schematic should return an error", func() {
			applyOpts.OamFiles = []string{
				"schematics/manually-scaled-frontend.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(MatchError(HavePrefix("Application configuration refers to component nginx-replicated, but no file provided the component schematic")))
		})
	})

	Context("Generate templates", func() {
		BeforeAll(func() {
			os.Chdir("../templates")
		})

		BeforeEach(func() {
			applyOpts = cli.NewApplyOpts()
			applyOpts.DryRun = true

			// Tests should not actually be calling AWS APIs, so give garbage credentials for the session
			session := session.Must(session.NewSessionWithOptions(
				session.Options{
					Config: aws.Config{
						Credentials: credentials.NewStaticCredentials("GARBAGE_ACCESS_KEY_ID", "GARBASE_SECRET_ACCESS_KEY", ""),
						Region:      aws.String("garbage-region-1"),
					},
				},
			))
			applyOpts.ComponentDeployer = cloudformation.New(session)
		})

		It("simple single worker component and configuration", func() {
			applyOpts.OamFiles = []string{
				"../integ-tests/schematics/worker.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(BeNil())

			actualTemplate, _ := filepath.Abs("oam-ecs-dry-run-results/oam-ecs-simple-worker-web-front-end-template.yaml")
			expectedTemplate, _ := filepath.Abs("../integ-tests/schematics/worker.expected.yaml")
			Expect(actualTemplate).Should(BeAnExistingFile())
			Expect(actualTemplate).Should(MatchCloudFormationTemplate(expectedTemplate))
		})

		It("simple single server component and configuration", func() {
			applyOpts.OamFiles = []string{
				"../integ-tests/schematics/nginx.yaml",
				"../integ-tests/schematics/manually-scaled-frontend.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(BeNil())

			actualTemplate, _ := filepath.Abs("oam-ecs-dry-run-results/oam-ecs-manual-scaler-app-web-front-end-template.yaml")
			expectedTemplate, _ := filepath.Abs("../integ-tests/schematics/manually-scaled-frontend.expected.yaml")
			Expect(actualTemplate).Should(BeAnExistingFile())
			Expect(actualTemplate).Should(MatchCloudFormationTemplate(expectedTemplate))
		})

		It("order of the files does not matter", func() {
			applyOpts.OamFiles = []string{
				"../integ-tests/schematics/manually-scaled-frontend.yaml",
				"../integ-tests/schematics/nginx.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(BeNil())

			actualTemplate, _ := filepath.Abs("oam-ecs-dry-run-results/oam-ecs-manual-scaler-app-web-front-end-template.yaml")
			expectedTemplate, _ := filepath.Abs("../integ-tests/schematics/manually-scaled-frontend.expected.yaml")
			Expect(actualTemplate).Should(BeAnExistingFile())
			Expect(actualTemplate).Should(MatchCloudFormationTemplate(expectedTemplate))
		})

		It("web server", func() {
			applyOpts.OamFiles = []string{
				"../integ-tests/schematics/webserver.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(BeNil())

			actualTemplate, _ := filepath.Abs("oam-ecs-dry-run-results/oam-ecs-webserver-app-web-front-end-template.yaml")
			expectedTemplate, _ := filepath.Abs("../integ-tests/schematics/webserver.frontend.expected.yaml")
			Expect(actualTemplate).Should(BeAnExistingFile())
			Expect(actualTemplate).Should(MatchCloudFormationTemplate(expectedTemplate))

			actualTemplate, _ = filepath.Abs("oam-ecs-dry-run-results/oam-ecs-webserver-app-backend-svc-template.yaml")
			expectedTemplate, _ = filepath.Abs("../integ-tests/schematics/webserver.backend.expected.yaml")
			Expect(actualTemplate).Should(BeAnExistingFile())
			Expect(actualTemplate).Should(MatchCloudFormationTemplate(expectedTemplate))
		})

		It("twitter bot", func() {
			applyOpts.OamFiles = []string{
				"../integ-tests/schematics/twitter-bot.yaml",
			}
			err := applyOpts.Execute()
			Expect(err).Should(BeNil())

			actualTemplate, _ := filepath.Abs("oam-ecs-dry-run-results/oam-ecs-twitter-bot-web-front-end-template.yaml")
			expectedTemplate, _ := filepath.Abs("../integ-tests/schematics/twitter-bot.frontend.expected.yaml")
			Expect(actualTemplate).Should(BeAnExistingFile())
			Expect(actualTemplate).Should(MatchCloudFormationTemplate(expectedTemplate))

			actualTemplate, _ = filepath.Abs("oam-ecs-dry-run-results/oam-ecs-twitter-bot-backend-svc-template.yaml")
			expectedTemplate, _ = filepath.Abs("../integ-tests/schematics/twitter-bot.backend.expected.yaml")
			Expect(actualTemplate).Should(BeAnExistingFile())
			Expect(actualTemplate).Should(MatchCloudFormationTemplate(expectedTemplate))
		})
	})
})
