package test

import (
	"flag"
	"fmt"
	"os"

	//regexp"
	"sort"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	//"github.com/stretchr/testify/require"
	"github.com/thedevsaddam/gojsonq/v2"
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
	var certificateARN string
	var vpcID string

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
	moddirs["0-gitlabCertificate"] = "../global/certificates/gitlab"
	moddirs["0-gitlabKeyPair"] = "../reg-primary/keypairs/gitlab"
	moddirs["1-customICEndpointSG"] = "../reg-primary/sgs/custom-ice"
	moddirs["1-gitlabSG"] = "../reg-primary/sgs/gitlab"
	moddirs["2-customICEndpoint"] = "../reg-primary/ices/custom"
	moddirs["2-gitlabInstance"] = "../reg-primary/instances/gitlab"
	moddirs["3-gitlabALB"] = "../global/albs/gitlab"

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

			// Query the json string representing state returned by 'terraform.Show'
			modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
				Where("address", "eq", "aws_vpc.this[0]").
				Select("values")
			// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
			values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

			// Make sure tags are applied
			for tag, content := range env["labels"].(map[string]interface{}) {
				if assert.Equal(t, values.(map[string]interface{})["tags"].(map[string]interface{})[tag].(string), content.(string)) {
					t.Logf("Tag check for tag '%s' PASSED for module in %s", tag, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Tag check for tag '%s' FAILED for module in %s, expected %s.", tag, terraformOptions.TerraformDir, content.(string))
				}
			}

			// Store the VPC ID
			vpcID = outputs["vpc_id"].(string)

		// GitLab Certificate module
		case "0-gitlabCertificate":
			// Make sure that prevent_destroy is set to false
			if assert.Contains(t, hclstring, "prevent_destroy = false") {
				t.Logf("Prevent destroy check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Prevent destroy check FAILED for module in %s, expected prevent_destroy to be set to false in terragrunt.hcl.", terraformOptions.TerraformDir)
			}

			// Make sure the resource name contains the prefix, environment and name
			if assert.Contains(t, outputs["distinct_domain_names"], inputs["name"].(string)+"."+env["dns"].(map[string]interface{})["domain"].(string)) {
				t.Logf("Distinct DNS name check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Distinct DNS name check FAILED for module in %s, expected %s.", terraformOptions.TerraformDir, inputs["name"].(string)+"."+env["dns"].(map[string]interface{})["domain"].(string))
			}

			certificateARN = outputs["acm_certificate_arn"].(string)

			// Query the json string representing state returned by 'terraform.Show'
			modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
				Where("address", "eq", "aws_acm_certificate.this[0]").
				Select("values")
			// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
			values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

			// Make sure tags are applied
			for tag, content := range env["labels"].(map[string]interface{}) {
				if assert.Equal(t, values.(map[string]interface{})["tags"].(map[string]interface{})[tag].(string), content.(string)) {
					t.Logf("Tag check for tag '%s' PASSED for module in %s", tag, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Tag check for tag '%s' FAILED for module in %s, expected %s.", tag, terraformOptions.TerraformDir, content.(string))
				}
			}

		// Custom Instance Connect Endpoint Security Group module
		case "1-customICEndpointSG":
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

			// Make sure the security group is assigned to the correct VPC
			if assert.Equal(t, vpcID, outputs["security_group_vpc_id"].(string)) {
				t.Logf("VPC ID check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("VPC ID check FAILED for module in %s, expected VPC ID to be %s", terraformOptions.TerraformDir, vpcID)
			}

			// Make sure the security group has the correct ingress rules
			cidr_blocks := inputs["ingress_cidr_blocks"].([]interface{})
			for i := range inputs["ingress_rules"].([]interface{}) {
				// Query the json string representing state returned by 'terraform.Show'
				modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
					Where("address", "eq", fmt.Sprintf("aws_security_group_rule.ingress_rules[%d]", i)).
					Select("values")
				// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
				values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

				// Compare the ingress rules with configured inputs
				rule := strings.Split(inputs["ingress_rules"].([]interface{})[i].(string), "-")
				if assert.Equal(t, strings.ToUpper(rule[0]), values.(map[string]interface{})["description"].(string)) &&
					assert.Equal(t, rule[len(rule)-1], values.(map[string]interface{})["protocol"].(string)) &&
					assert.Equal(t, cidr_blocks, values.(map[string]interface{})["cidr_blocks"]) {
					t.Logf("Ingress rule %d check PASSED for module in %s", i, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Ingress rule %d check FAILED for module in %s, expected CIDR to be %s, description to be %s and protocol to be %s", i, terraformOptions.TerraformDir, cidr_blocks, strings.ToUpper(rule[0]), rule[len(rule)-1])
				}
			}

			// Make sure there's an egress rule for SSH, or the endpoint won't connect
			if assert.GreaterOrEqual(t, len(inputs["egress_cidr_blocks"].([]interface{})), 1) &&
				assert.Contains(t, inputs["egress_rules"].([]interface{}), "ssh-tcp") {
				t.Logf("Egress rule 'ssh-tcp' check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Egress rule 'ssh-tcp' check FAILED for module in %s, expected an 'ssh-tcp' egress rule to be configured.", terraformOptions.TerraformDir)
			}

			// Make sure the security group has the correct egress rules
			cidr_blocks = inputs["egress_cidr_blocks"].([]interface{})
			for i := range inputs["egress_rules"].([]interface{}) {
				// Query the json string representing state returned by 'terraform.Show'
				modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
					Where("address", "eq", fmt.Sprintf("aws_security_group_rule.egress_rules[%d]", i)).
					Select("values")
				// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
				values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

				// Compare the egress rules with configured inputs
				rule := strings.Split(inputs["egress_rules"].([]interface{})[i].(string), "-")
				if assert.Equal(t, strings.ToUpper(rule[0]), values.(map[string]interface{})["description"].(string)) &&
					assert.Equal(t, rule[len(rule)-1], values.(map[string]interface{})["protocol"].(string)) &&
					assert.Equal(t, cidr_blocks, values.(map[string]interface{})["cidr_blocks"]) {
					t.Logf("Egress rule %d check PASSED for module in %s", i, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Egress rule %d check FAILED for module in %s, expected CIDR to be %s, description to be %s and protocol to be %s", i, terraformOptions.TerraformDir, cidr_blocks, strings.ToUpper(rule[0]), rule[len(rule)-1])
				}
			}

			// Query the json string representing state returned by 'terraform.Show'
			modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
				Where("address", "eq", "aws_security_group.this_name_prefix[0]").
				Select("values")
			// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
			values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

			// Make sure tags are applied
			for tag, content := range env["labels"].(map[string]interface{}) {
				if assert.Equal(t, values.(map[string]interface{})["tags"].(map[string]interface{})[tag].(string), content.(string)) {
					t.Logf("Tag check for tag '%s' PASSED for module in %s", tag, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Tag check for tag '%s' FAILED for module in %s, expected %s.", tag, terraformOptions.TerraformDir, content.(string))
				}
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

			// Make sure the security group is assigned to the correct VPC
			if assert.Equal(t, vpcID, outputs["security_group_vpc_id"].(string)) {
				t.Logf("VPC ID check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("VPC ID check FAILED for module in %s, expected VPC ID to be %s", terraformOptions.TerraformDir, vpcID)
			}

			// Make sure the security group has the correct ingress rules
			cidr_blocks := append(inputs["ingress_cidr_blocks"].([]interface{}), env["dependencies"].(map[string]interface{})["custom_vpc_mock_outputs"].(map[string]interface{})["vpc_cidr_block"].(string))
			for i := range inputs["ingress_rules"].([]interface{}) {
				// Query the json string representing state returned by 'terraform.Show'
				modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
					Where("address", "eq", fmt.Sprintf("aws_security_group_rule.ingress_rules[%d]", i)).
					Select("values")
				// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
				values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

				// Compare the ingress rules with configured inputs
				rule := strings.Split(inputs["ingress_rules"].([]interface{})[i].(string), "-")
				if assert.Equal(t, strings.ToUpper(rule[0]), values.(map[string]interface{})["description"].(string)) &&
					assert.Equal(t, rule[len(rule)-1], values.(map[string]interface{})["protocol"].(string)) &&
					assert.Equal(t, cidr_blocks, values.(map[string]interface{})["cidr_blocks"]) {
					t.Logf("Ingress rule %d check PASSED for module in %s", i, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Ingress rule %d check FAILED for module in %s, expected CIDR to be %s, description to be %s and protocol to be %s", i, terraformOptions.TerraformDir, cidr_blocks, strings.ToUpper(rule[0]), rule[len(rule)-1])
				}
			}

			// Make sure the security group has the correct egress rules
			cidr_blocks = inputs["egress_cidr_blocks"].([]interface{})
			for i := range inputs["egress_rules"].([]interface{}) {
				// Query the json string representing state returned by 'terraform.Show'
				modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
					Where("address", "eq", fmt.Sprintf("aws_security_group_rule.egress_rules[%d]", i)).
					Select("values")
				// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
				values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

				// Compare the egress rules with configured inputs
				rule := strings.Split(inputs["egress_rules"].([]interface{})[i].(string), "-")
				if assert.Equal(t, strings.ToUpper(rule[0]), values.(map[string]interface{})["description"].(string)) &&
					assert.Equal(t, rule[len(rule)-1], values.(map[string]interface{})["protocol"].(string)) &&
					assert.Equal(t, cidr_blocks, values.(map[string]interface{})["cidr_blocks"]) {
					t.Logf("Egress rule %d check PASSED for module in %s", i, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Egress rule %d check FAILED for module in %s, expected CIDR to be %s, description to be %s and protocol to be %s", i, terraformOptions.TerraformDir, cidr_blocks, strings.ToUpper(rule[0]), rule[len(rule)-1])
				}
			}

			// Query the json string representing state returned by 'terraform.Show'
			modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
				Where("address", "eq", "aws_security_group.this_name_prefix[0]").
				Select("values")
			// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
			values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

			// Make sure tags are applied
			for tag, content := range env["labels"].(map[string]interface{}) {
				if assert.Equal(t, values.(map[string]interface{})["tags"].(map[string]interface{})[tag].(string), content.(string)) {
					t.Logf("Tag check for tag '%s' PASSED for module in %s", tag, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Tag check for tag '%s' FAILED for module in %s, expected %s.", tag, terraformOptions.TerraformDir, content.(string))
				}
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

			// Query the json string representing state returned by 'terraform.Show'
			modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
				Where("address", "eq", "aws_key_pair.this[0]").
				Select("values")
			// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
			values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

			// Make sure tags are applied
			for tag, content := range env["labels"].(map[string]interface{}) {
				if assert.Equal(t, values.(map[string]interface{})["tags"].(map[string]interface{})[tag].(string), content.(string)) {
					t.Logf("Tag check for tag '%s' PASSED for module in %s", tag, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Tag check for tag '%s' FAILED for module in %s, expected %s.", tag, terraformOptions.TerraformDir, content.(string))
				}
			}

		// Custom Instance Connect Endpoint module
		case "2-customICEndpoint":
			// Make sure that prevent_destroy is set to false
			if assert.Contains(t, hclstring, "prevent_destroy = false") {
				t.Logf("Prevent destroy check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Prevent destroy check FAILED for module in %s, expected prevent_destroy to be set to false in terragrunt.hcl.", terraformOptions.TerraformDir)
			}

			// Make sure the endpoint is in the correct zone
			if assert.Equal(t, outputs["availability_zone"].(string), pregion["location"].(string)+pregion["zone_preference"].(string)) {
				t.Logf("Availability zone check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Availability zone check FAILED for module in %s, expected instance to be deployed in zone %s", terraformOptions.TerraformDir, pregion["location"].(string)+pregion["zone_preference"].(string))
			}

			// Make sure the endpoint configures client IP preservation correctly
			if assert.Equal(t, outputs["preserve_client_ip"].(bool), inputs["preserve_client_ip"].(bool)) {
				t.Logf("Preserve client IP check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Preserve client IP check FAILED for module in %s, expected preserve client IP to be %t", terraformOptions.TerraformDir, inputs["preserve_client_ip"].(bool))
			}

			// Make sure tags are applied
			for tag, content := range env["labels"].(map[string]interface{}) {
				if assert.Equal(t, outputs["tags"].(map[string]interface{})[tag].(string), content.(string)) {
					t.Logf("Tag check for tag '%s' PASSED for module in %s", tag, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Tag check for tag '%s' FAILED for module in %s, expected %s.", tag, terraformOptions.TerraformDir, content.(string))
				}
			}

		// GitLab EC2 Instance module
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

			// Make sure the instance is in the correct zone
			if assert.Equal(t, outputs["availability_zone"].(string), pregion["location"].(string)+pregion["zone_preference"].(string)) {
				t.Logf("Availability zone check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Availability zone check FAILED for module in %s, expected instance to be deployed in zone %s", terraformOptions.TerraformDir,
					pregion["location"].(string)+pregion["zone_preference"].(string))
			}

			// Make sure the instance does not configure a public IP address
			if assert.False(t, inputs["public_ip"].(bool)) || assert.Empty(t, outputs["public_ip"]) {
				t.Logf("Public IP not assigned check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Public IP not assigned check FAILED for module in %s, expected no public IP address association to be configured.", terraformOptions.TerraformDir)
			}

			// Query the json string representing state returned by 'terraform.Show'
			modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
				Where("address", "eq", "aws_instance.this[0]").
				Select("values")
			// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
			values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

			// Make sure tags are applied
			for tag, content := range env["labels"].(map[string]interface{}) {
				if assert.Equal(t, values.(map[string]interface{})["tags"].(map[string]interface{})[tag].(string), content.(string)) {
					t.Logf("Tag check for tag '%s' PASSED for module in %s", tag, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Tag check for tag '%s' FAILED for module in %s, expected %s.", tag, terraformOptions.TerraformDir, content.(string))
				}
			}

		// GitLab Applicaton Load Balancer module
		case "3-gitlabALB":
			// Make sure that prevent_destroy is set to false
			if assert.Contains(t, hclstring, "prevent_destroy = false") {
				t.Logf("Prevent destroy check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Prevent destroy check FAILED for module in %s, expected prevent_destroy to be set to false in terragrunt.hcl.", terraformOptions.TerraformDir)
			}

			// Make sure the fqdn contains the name and domain
			if assert.Contains(t, outputs["route53_records"].(map[string]interface{})[inputs["dns"].([]interface{})[0].(map[string]interface{})["name"].(string)].(map[string]interface{})["fqdn"].(string),
				inputs["dns"].([]interface{})[0].(map[string]interface{})["name"].(string)+"."+env["dns"].(map[string]interface{})["domain"].(string)) {
				t.Logf("Distinct DNS name check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Distinct DNS name check FAILED for module in %s, expected %s.", terraformOptions.TerraformDir,
					inputs["name"].(string)+"."+env["dns"].(map[string]interface{})["domain"].(string))
			}

			// Make sure the listener is on the correct port
			if assert.Equal(t, inputs["listeners"].(map[string]interface{})["https"].(map[string]interface{})["port"].(int),
				int(outputs["listeners"].(map[string]interface{})["https"].(map[string]interface{})["port"].(float64))) {
				t.Logf("Listener port check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Listener port check FAILED for module in %s, expected %d", terraformOptions.TerraformDir,
					inputs["listeners"].(map[string]interface{})["https"].(map[string]interface{})["port"].(int))
			}

			// Make sure the listener is using the correct protocol
			if assert.Equal(t, inputs["listeners"].(map[string]interface{})["https"].(map[string]interface{})["protocol"].(string),
				outputs["listeners"].(map[string]interface{})["https"].(map[string]interface{})["protocol"].(string)) {
				t.Logf("Listener protocol check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Listener protocol check FAILED for module in %s, expected %s", terraformOptions.TerraformDir, certificateARN)
			}

			// Make sure the correct certficate is linked to the ALB
			if assert.Equal(t, certificateARN, outputs["listeners"].(map[string]interface{})["https"].(map[string]interface{})["certificate_arn"].(string)) {
				t.Logf("Certificate ARN check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Certificate ARN check FAILED for module in %s, expected %s", terraformOptions.TerraformDir, certificateARN)
			}

			// Make sure the target group forwards to the correct port
			if assert.Equal(t, inputs["targets"].(map[string]interface{})["gitlab"].(map[string]interface{})["port"].(int),
				int(outputs["target_groups"].(map[string]interface{})["gitlab"].(map[string]interface{})["port"].(float64))) {
				t.Logf("Target port check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Target port check FAILED for module in %s, expected %d", terraformOptions.TerraformDir,
					inputs["targets"].(map[string]interface{})["gitlab"].(map[string]interface{})["port"].(int))
			}

			// Make sure the target group forwards using the correct protocol
			if assert.Equal(t, inputs["targets"].(map[string]interface{})["gitlab"].(map[string]interface{})["protocol"].(string),
				outputs["target_groups"].(map[string]interface{})["gitlab"].(map[string]interface{})["protocol"].(string)) {
				t.Logf("Target protocol check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Target protocol check FAILED for module in %s, expected %s", terraformOptions.TerraformDir,
					inputs["targets"].(map[string]interface{})["gitlab"].(map[string]interface{})["protocol"].(string))
			}

			// Make sure the target group configures the correct health check URI
			path_exists := false
			for _, object := range outputs["target_groups"].(map[string]interface{})["gitlab"].(map[string]interface{})["health_check"].([]interface{}) {
				if assert.Equal(t, object.(map[string]interface{})["path"].(string),
					inputs["targets"].(map[string]interface{})["gitlab"].(map[string]interface{})["health_check"].(map[string]interface{})["path"].(string)) {
					path_exists = true
				}
			}
			if path_exists {
				t.Logf("Health check URI check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Health check URI check FAILED for module in %s, expected %s", terraformOptions.TerraformDir,
					inputs["targets"].(map[string]interface{})["gitlab"].(map[string]interface{})["health_check"].(map[string]interface{})["path"].(string))
			}

			// Query the json string representing state returned by 'terraform.Show'
			modulejson := gojsonq.New().JSONString(terraform.Show(t, terraformOptions)).From("values.root_module.resources").
				Where("address", "eq", "aws_lb.this[0]").
				Select("values")
			// Execute the above query; since it modifies the pointer we can only do this once, so we add it to a variable
			values := modulejson.Get().([]interface{})[0].(map[string]interface{})["values"]

			// Make sure deletion protection is set properly
			if assert.Equal(t, inputs["deletion_protection"].(bool), values.(map[string]interface{})["enable_deletion_protection"].(bool)) {
				t.Logf("Deletion protection check PASSED for module in %s", terraformOptions.TerraformDir)
			} else {
				t.Errorf("Deletion protection check FAILED for module in %s, expected %t.", terraformOptions.TerraformDir, inputs["deletion_protection"].(bool))
			}

			// Make sure tags are applied
			for tag, content := range env["labels"].(map[string]interface{}) {
				if assert.Equal(t, values.(map[string]interface{})["tags"].(map[string]interface{})[tag].(string), content.(string)) {
					t.Logf("Tag check for tag '%s' PASSED for module in %s", tag, terraformOptions.TerraformDir)
				} else {
					t.Errorf("Tag check for tag '%s' FAILED for module in %s, expected %s.", tag, terraformOptions.TerraformDir, content.(string))
				}
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
