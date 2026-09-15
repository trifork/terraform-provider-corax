default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && TF_ACC=1 go test -v -cover -timeout 120m ./...; \
	else \
		TF_ACC=1 go test -v -cover -timeout 120m ./...; \
	fi

# --- Local integration harness (test/integration) -----------------------------
# Builds the provider into ./bin, points Terraform at it with a dev_overrides
# CLI config, and runs plan/apply/destroy against a real Corax environment.
# Requires CORAX_API_ENDPOINT and CORAX_API_KEY (a .env file is sourced if present).

INTEGRATION_DIR := $(CURDIR)/test/integration
INTEGRATION_BIN := $(CURDIR)/bin
TFRC            := $(INTEGRATION_DIR)/dev.tfrc
TFVARS          := $(INTEGRATION_DIR)/run.auto.tfvars

integration-setup:
	@mkdir -p $(INTEGRATION_BIN)
	go build -o $(INTEGRATION_BIN)/terraform-provider-corax .
	@printf 'provider_installation {\n  dev_overrides {\n    "registry.terraform.io/trifork/corax" = "%s"\n  }\n  direct {}\n}\n' "$(INTEGRATION_BIN)" > $(TFRC)
	@test -f $(TFVARS) || printf 'run_id = "%s"\n' "$$(openssl rand -hex 4)" > $(TFVARS)
	@echo "run_id: $$(cat $(TFVARS))"

integration-plan: integration-setup
	@$(INTEGRATION_ENV) terraform -chdir=$(INTEGRATION_DIR) plan

integration-apply: integration-setup
	@$(INTEGRATION_ENV) terraform -chdir=$(INTEGRATION_DIR) apply -auto-approve

integration-destroy: integration-setup
	@$(INTEGRATION_ENV) terraform -chdir=$(INTEGRATION_DIR) destroy -auto-approve
	@rm -f $(TFVARS)

# Sources .env (if present) and exports the dev_overrides CLI config.
INTEGRATION_ENV = set -a; [ -f $(CURDIR)/.env ] && . $(CURDIR)/.env; set +a; TF_CLI_CONFIG_FILE=$(TFRC)

.PHONY: fmt lint test testacc build install generate integration-setup integration-plan integration-apply integration-destroy
