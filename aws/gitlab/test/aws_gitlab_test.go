package test

import (
	"flag"
	"fmt"
	"os"

	//regexp"
	"sort"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	//"github.com/stretchr/testify/require"
	//"github.com/thedevsaddam/gojsonq/v2"
	"gopkg.in/yaml.v3"
)

// Flag to destroy the target environment after tests
var destroy = flag.Bool("destroy", false, "destroy environment after tests")

func TestTerragruntDeployment(t *testing.T) {

	// Terraform options
	binary := "terragrunt"
	rootdir := "../."
	moddirs := make(map[string]string)

	// Non-local vars to evaluate state between modules

	// Reusable vars for unmarshalling YAML files
	var err error
	var yfile []byte

	// Define the deployment root
	terraformDeploymentOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir:    rootdir,
		TerraformBinary: binary,
	})

	// Check for standard global configuration files
	if !fileExists(terraformDeploymentOptions.TerraformDir + "/env.yaml") {
		t.Error("Configuration check FAILED. Environment configuration file not found!")
	}
	if !fileExists(terraformDeploymentOptions.TerraformDir + "/../local.aws.yaml") {
		if !fileExists(terraformDeploymentOptions.TerraformDir + "/../aws.yaml") {
			t.Error("Configuration check FAILED. Platform configuration file not found!")
		}
	}
	if !fileExists(terraformDeploymentOptions.TerraformDir + "/reg-primary/region.yaml") {
		t.Error("Configuration check FAILED. Primary region configuration file not found!")
	}
	if !fileExists(terraformDeploymentOptions.TerraformDir + "/reg-secondary/region.yaml") {
		t.Error("Configuration check FAILED. Secondary region configuration file not found!")
	}
	if !fileExists(terraformDeploymentOptions.TerraformDir + "/versions.yaml") {
		t.Error("Configuration check FAILED. Versions configuration file not found!")
	}

	// Define modules
	moddirs["0-customVPC"] = "../reg-primary/vpcs/custom"
	moddirs["1-gitlabSG"] = "../reg-primary/sgs/gitlab"
	moddirs["1-gitlabKeyPair"] = "../reg-primary/keypairs/gitlab"
	moddirs["2-gitlabInstance"] = "../reg-primary/instances/gitlab"

	// Maps are unsorted, so sort the keys to process the modules in order
	modkeys := make([]string, 0, len(moddirs))
	for k := range moddirs {
		modkeys = append(modkeys, k)
	}
	sort.Strings(modkeys)

	for _, module := range modkeys {
		terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
			TerraformDir:    moddirs[module],
			TerraformBinary: binary,
		})

		fmt.Println("Validating module:", module)

		// Sanity test
		terraform.Validate(t, terraformOptions)

		// Check for standard files
		if !fileExists(terraformOptions.TerraformDir + "/inputs.yaml") {
			t.Errorf("Module configuration check FAILED. Inputs file not found in %s", terraformOptions.TerraformDir)
		}
		if !fileExists(terraformOptions.TerraformDir + "/remotestate.tf") {
			t.Errorf("Module configuration check FAILED. Remote state file not found in %s", terraformOptions.TerraformDir)
		}
		if !fileExists(terraformOptions.TerraformDir + "/terragrunt.hcl") {
			t.Errorf("Module configuration check FAILED. Terragrunt configuration file not found in %s", terraformOptions.TerraformDir)
		}
	}

	// Read and store the env.yaml
	yfile, err = os.ReadFile(terraformDeploymentOptions.TerraformDir + "/env.yaml")
	if err != nil {
		t.Error("Configuration check FAILED. Could not read environment configuration file!")
	}

	env := make(map[string]interface{})
	err = yaml.Unmarshal(yfile, &env)
	if err != nil {
		t.Error("Configuration check FAILED. Could not parse environment configuration file!")
	}

	// Read and store the aws.yaml
	if fileExists(terraformDeploymentOptions.TerraformDir + "/../local.aws.yaml") {
		yfile, err = os.ReadFile(terraformDeploymentOptions.TerraformDir + "/../local.aws.yaml")
		if err != nil {
			t.Error("Configuration check FAILED. Could not read platform configuration file!")
		}
	} else {
		yfile, err = os.ReadFile(terraformDeploymentOptions.TerraformDir + "/../aws.yaml")
		if err != nil {
			t.Error("Configuration check FAILED. Could not read platform configuration file!")
		}
	}

	platform := make(map[string]interface{})
	err = yaml.Unmarshal(yfile, &platform)
	if err != nil {
		t.Error("Configuration check FAILED. Could not parse platform configuration file!")
	}

	// Read and store the reg-primary/region.yaml
	yfile, err = os.ReadFile(terraformDeploymentOptions.TerraformDir + "/reg-primary/region.yaml")
	if err != nil {
		t.Error("Configuration check FAILED. Could not read primary region configuration file!")
	}

	pregion := make(map[string]interface{})
	err = yaml.Unmarshal(yfile, &pregion)
	if err != nil {
		t.Error("Configuration check FAILED. Could not parse primary region configuration file!")
	}

	// Read and store the reg-secondary/region.yaml
	yfile, err = os.ReadFile(terraformDeploymentOptions.TerraformDir + "/reg-secondary/region.yaml")
	if err != nil {
		t.Error("Configuration check FAILED. Could not read secondary region configuration file!")
	}

	sregion := make(map[string]interface{})
	err = yaml.Unmarshal(yfile, &sregion)
	if err != nil {
		t.Error("Configuration check FAILED. Could not parse secondary region configuration file!")
	}

	// Read and store the versions.yaml
	yfile, err = os.ReadFile(terraformDeploymentOptions.TerraformDir + "/versions.yaml")
	if err != nil {
		t.Error("Configuration check FAILED. Could not read versions configuration file!")
	}

	versions := make(map[string]interface{})
	err = yaml.Unmarshal(yfile, &versions)
	if err != nil {
		t.Error("Configuration check FAILED. Could not parse versions configuration file!")
	}

	// Clean up after ourselves if flag is set
	if *destroy {
		defer terraform.TgDestroyAll(t, terraformDeploymentOptions)
	}

	// Deploy the composition
	terraform.TgApplyAll(t, terraformDeploymentOptions)

	for _, module := range modkeys {
		terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
			TerraformDir:    moddirs[module],
			TerraformBinary: binary,
		})

		fmt.Println("Testing module:", module)

		// Read the provider output and verify configured version
		providers := terraform.RunTerraformCommand(t, terraformOptions, terraform.FormatArgs(terraformOptions, "providers")...)
		if assert.Contains(t, providers, "provider[registry.terraform.io/hashicorp/aws] ~> "+versions["aws_provider_version"].(string)) {
			t.Logf("AWS provider version check PASSED for module in %s", terraformOptions.TerraformDir)
		} else {
			t.Errorf("AWS provider version check FAILED for module in %s, expected %s", terraformOptions.TerraformDir,
				"provider[registry.terraform.io/hashicorp/aws] ~> "+versions["aws_provider_version"].(string))
		}

		// Read the inputs.yaml
		yfile, err := os.ReadFile(terraformOptions.TerraformDir + "/inputs.yaml")
		if err != nil {
			t.Errorf("Module configuration check FAILED for module in %s. Could not read inputs.yaml file!", terraformOptions.TerraformDir)
		}

		inputs := make(map[string]interface{})
		err = yaml.Unmarshal(yfile, &inputs)
		if err != nil {
			t.Errorf("Module configuration check FAILED for module in %s. Could not parse inputs.yaml file!", terraformOptions.TerraformDir)
		}

		// Read the terragrunt.hcl
		hclfile, err := os.ReadFile(terraformOptions.TerraformDir + "/terragrunt.hcl")
		if err != nil {
			t.Errorf("Module configuration check FAILED for module in %s. Could not read terragrunt.hcl file!", terraformOptions.TerraformDir)
		}

		hclstring := string(hclfile)

		// Make sure the path referes to the correct parent hcl file
		if assert.Contains(t, hclstring, "path = find_in_parent_folders(\"aws_gitlab_terragrunt.hcl\")") {
			t.Logf("Parent terragrunt file check PASSED for module in %s", terraformOptions.TerraformDir)
		} else {
			t.Errorf("Parent terragrunt file check FAILED for module in %s, expected %s", terraformOptions.TerraformDir,
				"path = find_in_parent_folders(\"aws_gitlab_terragrunt.hcl\").")
		}

		// Collect the outputs
		outputs := terraform.OutputAll(t, terraformOptions)

		if assert.NotEmpty(t, outputs) {
			t.Logf("Terragrunt outputs check PASSED for module in %s", terraformOptions.TerraformDir)
		} else {
			t.Errorf("Terragrunt outputs check FAILED for module in %s, expected outputs not to be empty.", terraformOptions.TerraformDir)
		}

		// Add module-specific tests below
		// Remember that we're in a loop, so group tests by module name (modules range keys)
		// The following collections are available for tests:
		//   platform, env, pregion, sregion, versions, inputs, outputs
		// Two key patterns are available.
		// 1. Reference the output map returned by terraform.OutputAll (ie. the output of "terragrunt output")
		//		require.Equal(t, pregion["location"], outputs["location"])
		// 2. Query the json string representing state returned by terraform.Show (ie. the output of "terragrunt show -json")
		//		modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
		//			Where("address", "eq", "resource.this").
		//			Select("values")
		//		// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
		//		values := modulejson.Get()

		// Module-specific tests
		switch module {

		// Custom VPC module
		case "0-customVPC":
			// Make sure that prevent_destroy is set to false
			if assert.Contains(t, hclstring, "prevent_destroy = false") {
				t.Logf("Prevent destroy check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Prevent destroy check FAILED for module in %s, expected prevent_destroy to be set to false in terragrunt.hcl.", terraformOptions.TerraformDir)
			}

			// Make sure the resource name contains the prefix, environment and name
			if (assert.Contains(t, outputs["name"], platform["prefix"].(string))) &&
				(assert.Contains(t, outputs["name"], env["environment"].(string))) &&
				(assert.Contains(t, outputs["name"], inputs["name"].(string))) {
				t.Logf("Resource name check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Resource name check FAILED for module in %s, expected name to contain configured prefix, environment and name elements.", terraformOptions.TerraformDir)
			}

			// Make sure there is an Internet gateway deployed
			if assert.NotEmpty(t, outputs["igw_id"]) {
				t.Logf("Internet gateway check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Internet gateway check FAILED for module in %s, expected Internet gateway to be deployed.", terraformOptions.TerraformDir)
			}

			// Make sure DNS is enabled
			if assert.True(t, outputs["vpc_enable_dns_support"].(bool)) {
				t.Logf("DNS check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("DNS check FAILED for module in %s, expected DNS to be enabled.", terraformOptions.TerraformDir)
			}

			// Make sure the correct CIDR block is configured
			if assert.Equal(t, inputs["cidr"].(string), outputs["vpc_cidr_block"].(string)) {
				t.Logf("CIDR block check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("CIDR block check FAILED for module in %s, expected configured CIDR block to be %s.", terraformOptions.TerraformDir, inputs["cidr"].(string))
			}

			// Make sure the resources are deployed to the appropriate zones
			for _, z := range inputs["zones"].([]interface{}) {
				if assert.Contains(t, outputs["azs"], pregion["location"].(string)+z.(string)) {
					t.Logf("Zone deployment check for zone %s PASSED for module in %s", pregion["location"].(string)+z.(string), terraformOptions.TerraformDir)
				} else {
					t.Errorf("Zone deployment check FAILED for module in %s, expected deployment to zone %s.", terraformOptions.TerraformDir, pregion["location"].(string)+z.(string))
				}
			}

			// Make sure there is one NAT gateway per zone
			if assert.Equal(t, len(outputs["azs"].([]interface{})), len(outputs["natgw_ids"].([]interface{}))) {
				t.Logf("NAT gateway check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("NAT gateway check FAILED for module in %s, expected one NAT gateway per zone.", terraformOptions.TerraformDir)
			}

		// GitLab Security Group module
		case "1-gitlabSG":
			// Make sure that prevent_destroy is set to false
			if assert.Contains(t, hclstring, "prevent_destroy = false") {
				t.Logf("Prevent destroy check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Prevent destroy check FAILED for module in %s, expected prevent_destroy to be set to false in terragrunt.hcl.", terraformOptions.TerraformDir)
			}

			// Make sure the resource name contains the prefix, environment and name
			if (assert.Contains(t, outputs["security_group_name"], platform["prefix"].(string))) &&
				(assert.Contains(t, outputs["security_group_name"], env["environment"].(string))) &&
				(assert.Contains(t, outputs["security_group_name"], inputs["name"].(string))) {
				t.Logf("Resource name check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Resource name check FAILED for module in %s, expected name to contain configured prefix, environment and name elements.", terraformOptions.TerraformDir)
			}

		// GitLab Key Pair module
		case "1-gitlabKeyPair":
			// Make sure that prevent_destroy is set to false
			if assert.Contains(t, hclstring, "prevent_destroy = false") {
				t.Logf("Prevent destroy check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Prevent destroy check FAILED for module in %s, expected prevent_destroy to be set to false in terragrunt.hcl.", terraformOptions.TerraformDir)
			}

			// Make sure the resource name contains the prefix, environment and name
			if (assert.Contains(t, outputs["key_pair_name"], platform["prefix"].(string))) &&
				(assert.Contains(t, outputs["key_pair_name"], env["environment"].(string))) &&
				(assert.Contains(t, outputs["key_pair_name"], inputs["name"].(string))) {
				t.Logf("Resource name check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Resource name check FAILED for module in %s, expected name to contain configured prefix, environment and name elements.", terraformOptions.TerraformDir)
			}

		// GitLab Security Group module
		case "2-gitlabInstance":
			// Make sure that prevent_destroy is set to false
			if assert.Contains(t, hclstring, "prevent_destroy = false") {
				t.Logf("Prevent destroy check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Prevent destroy check FAILED for module in %s, expected prevent_destroy to be set to false in terragrunt.hcl.", terraformOptions.TerraformDir)
			}

			// Make sure the resource name contains the prefix, environment and name
			if (assert.Contains(t, outputs["tags_all"].(map[string]interface{})["Name"], platform["prefix"].(string))) &&
				(assert.Contains(t, outputs["tags_all"].(map[string]interface{})["Name"], env["environment"].(string))) &&
				(assert.Contains(t, outputs["tags_all"].(map[string]interface{})["Name"], inputs["name"].(string))) {
				t.Logf("Resource name check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Resource name check FAILED for module in %s, expected name to contain configured prefix, environment and name elements.", terraformOptions.TerraformDir)
			}
		}
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
