package test

import (
	"flag"
	"os"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// Flag to destroy the target environment after tests
var destroy = flag.Bool("destroy", false, "destroy environment after tests")

func TestAWSInstanceConnect(t *testing.T) {
	// Set execution directory
	terraformOptions := &terraform.Options{
		TerraformDir: "../fixtures/example1",
	}

	// Check for versions file
	if !assert.FileExists(t, terraformOptions.TerraformDir+"/../versions.yaml") {
		t.Fail()
	}

	// Read and store the versions.yaml
	yfile, err := os.ReadFile(terraformOptions.TerraformDir + "/../versions.yaml")
	if err != nil {
		t.Fail()
	}

	versions := make(map[string]interface{})
	err = yaml.Unmarshal(yfile, &versions)
	if err != nil {
		t.Fail()
	}

	// Read the version output and verify the configured version
	goversion := runtime.Version()

	if assert.GreaterOrEqual(t, goversion, "go"+versions["golang_runtime_version"].(string)) {
		t.Logf("Go runtime version check PASSED, expected version >= '%s', got '%s'", "go"+versions["golang_runtime_version"].(string), goversion)
	} else {
		t.Errorf("Go runtime version check FAILED, expected version >= '%s', got '%s'", "go"+versions["golang_runtime_version"].(string), goversion)
	}

	// Check for inputs file
	if !assert.FileExists(t, terraformOptions.TerraformDir+"/inputs.yaml") {
		t.Fail()
	}

	// Read and store the inputs.yaml
	yfile, err = os.ReadFile(terraformOptions.TerraformDir + "/inputs.yaml")
	if err != nil {
		t.Fail()
	}

	inputs := make(map[string]interface{})
	err = yaml.Unmarshal(yfile, &inputs)
	if err != nil {
		t.Fail()
	}

	// Initialize the deployment
	terraform.Init(t, terraformOptions)

	// Sanity test
	terraform.Validate(t, terraformOptions)

	// Read the version command output
	version := terraform.RunTerraformCommand(t, terraformOptions, terraform.FormatArgs(terraformOptions, "version")...)

	// Verify configured Terraform version
	if assert.Contains(t, version, "Terraform v"+versions["terraform_binary_version"].(string)) {
		t.Logf("Terraform version check PASSED, expected version '~> %s', got \n%s", versions["terraform_binary_version"].(string), version)
	} else {
		t.Errorf("Terraform version check FAILED, expected version '~> %s', got \n%s", versions["terraform_binary_version"].(string), version)
	}

	// Verify configured provider version
	if assert.Contains(t, version, "provider registry.terraform.io/hashicorp/aws v"+versions["aws_provider_version"].(string)) {
		t.Logf("Provider version check PASSED, expected hashicorp/aws version '~> %s', got \n%s", versions["aws_provider_version"].(string), version)
	} else {
		t.Errorf("Provider version check FAILED, expected hashicorp/aws version '~> %s', got \n%s", versions["aws_provider_version"].(string), version)
	}

	// Defer Terraform destroy only if flag is set
	if *destroy {
		defer terraform.Destroy(t, terraformOptions)
	}

	// Create resources
	terraform.Apply(t, terraformOptions)

	// Store outputs
	outputs := terraform.OutputAll(t, terraformOptions)

	// Test for valid output
	if assert.NotNil(t, outputs["subnet_id"]) {
		t.Logf("Subnet output test PASSED. Expected output to be string, got %s", outputs["subnet_id"].(string))
	} else {
		t.Error("Subnet output test FAILED. Expected output to be string, got nil")
	}

	if assert.NotNil(t, outputs["vpc_id"]) {
		t.Logf("VPC output test PASSED. Expected output to be string, got %s", outputs["vpc_id"].(string))
	} else {
		t.Error("VPC output test FAILED. Expected output to be string, got nil")
	}
}
