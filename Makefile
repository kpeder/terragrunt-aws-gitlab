.PHONY: help
help:
	@echo 'make <target>'
	@echo ''
	@echo 'Targets:'
	@echo '    help                 Show this help'
	@echo '    pre-commit           Run pre-commit checks'
	@echo ''
	@echo '    aws_gitlab_clean     Clean up state files'
	@echo '    aws_gitlab_configure Configure the deployment'
	@echo '    aws_gitlab_deploy    Deploy configured resources'
	@echo '    aws_gitlab_init      Initialize modules, providers'
	@echo '    aws_gitlab_install   Install Terraform, Terragrunt'
	@echo '    aws_gitlab_lint      Run Go linters'
	@echo '    aws_gitlab_plan      Show deployment plan'
	@echo '    aws_gitlab_test      Run deployment tests and clean up (CI loop)'
	@echo ''

.PHONY: pre-commit
pre-commit:
	@pre-commit run -a

.PHONY: aws_gitlab_clean
aws_gitlab_clean:
	@cd aws/gitlab && chmod +x ./scripts/prune_caches.sh && ./scripts/prune_caches.sh .
	@cd aws/gitlab/test && rm -f go.mod go.sum

.PHONY: aws_gitlab_configure
aws_gitlab_configure:
	@cd aws/gitlab && ./scripts/configure.sh -d bytecount.net -e demo -o kpeder -p us-east-2 -s us-west-2 -t devops -z Z2OCSN1ZPHG5PO

.PHONY: aws_gitlab_deploy
aws_gitlab_deploy: aws_gitlab_configure aws_gitlab_init
	@cd aws/gitlab/test && go test -v -timeout 20m

.PHONY: aws_gitlab_init
aws_gitlab_init: aws_gitlab_configure
	@cd aws/gitlab && terragrunt run-all init
	@cd aws/gitlab/test && go mod init aws_gitlab_test.go; go mod tidy

.PHONY: aws_gitlab_install
aws_gitlab_install:
	@chmod +x ./scripts/*.sh
	@sudo ./scripts/install_terraform.sh -v ./aws/gitlab/versions.yaml
	@sudo ./scripts/install_terragrunt.sh -v ./aws/gitlab/versions.yaml

.PHONY: aws_gitlab_lint
aws_gitlab_lint: aws_gitlab_configure aws_gitlab_init
	@cd aws/gitlab/test && golangci-lint run --print-linter-name --verbose aws_gitlab_test.go

.PHONY: aws_gitlab_plan
aws_gitlab_plan: aws_gitlab_configure aws_gitlab_init
	@cd aws/gitlab && terragrunt run-all plan

.PHONY: aws_gitlab_test
aws_gitlab_test: aws_gitlab_configure aws_gitlab_lint
	@cd aws/gitlab/test && go test -v -destroy -timeout 20m
