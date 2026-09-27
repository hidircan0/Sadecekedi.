ENDPOINT := http://localhost:4566
AWS := AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_DEFAULT_REGION=us-east-1 aws --endpoint-url=$(ENDPOINT)
TF_DIR := ansible/terraform
TF := terraform -chdir=$(TF_DIR)

.DEFAULT_GOAL := help

VENV := .venv
ANSIBLE := $(VENV)/bin/ansible-playbook
REPO := $(abspath $(CURDIR)/..)
export ANSIBLE_HOME := $(CURDIR)/.ansible
CONTRACT_TARGETS := -target=local_file.backend_env -target=local_file.validator_env -target=local_file.k8s_endpoints

.PHONY: help up init plan apply destroy status state down ansible ansible-validator deploy deploy-k3s deploy-gpu deploy-split up-app down-app forward

help:
	@echo "Usage:"
	@echo "  make up       Start LocalStack"
	@echo "  make init     terraform init (ansible/terraform)"
	@echo "  make plan     terraform plan"
	@echo "  make apply    terraform apply (writes ansible/terraform/generated/)"
	@echo "  make status   Query resources via AWS CLI at $(ENDPOINT)"
	@echo "  make state    terraform state list"
	@echo "  make destroy  terraform destroy"
	@echo "  make down     Stop LocalStack"
	@echo "  make ansible           Contract check (terraform files + playbook)"
	@echo "  make deploy            k3s varsa cluster, yoksa Docker Compose"
	@echo "  make up-app            Cluster'suz: postgres+minio+backend+validator"
	@echo "  make down-app          Compose stack'i kapat"
	@echo "  make deploy-k3s        Zorla k3s (k3s kurulu olmalı)"
	@echo "  make deploy-gpu        GPU validator → k3s"
	@echo "  make deploy-split      API cluster'da, validator dışarıda"
	@echo "  make forward           kubectl port-forward :8080"
	@echo "  make ansible-validator validator_ssh_host=IP"
	@echo ""
	@echo "Günlük hot-reload hâlâ: skaffold dev -n sadecekedi -p cpu"
	@echo "Bu sisteme geçiş: make deploy  (Skaffold silinmez)"

up:
	docker compose up -d
	@echo "Waiting for LocalStack health..."
	@until curl -sf $(ENDPOINT)/_localstack/health >/dev/null; do sleep 1; done
	@echo "LocalStack is ready at $(ENDPOINT)"

init:
	$(TF) init

plan:
	$(TF) plan

apply:
	mkdir -p $(TF_DIR)/generated
	$(TF) apply

destroy:
	$(TF) destroy

status:
	@echo "=== S3 ==="
	$(AWS) s3 ls
	@echo "=== SQS ==="
	$(AWS) sqs list-queues
	@echo "=== SNS ==="
	$(AWS) sns list-topics
	@echo "=== DynamoDB ==="
	$(AWS) dynamodb list-tables
	@echo "=== IAM roles (lab prefix) ==="
	$(AWS) iam list-roles --query 'Roles[?starts_with(RoleName, `localstack-lab`)].RoleName' --output text

state:
	$(TF) state list

down:
	docker compose down

$(VENV):
	python3 -m venv $(VENV)
	$(VENV)/bin/pip install -U pip ansible-core

ansible: $(VENV)
	mkdir -p $(TF_DIR)/generated
	$(TF) apply -auto-approve $(CONTRACT_TARGETS)
	cd ansible && ../$(ANSIBLE) site.yml

ansible-validator: $(VENV)
	cd ansible && ../$(ANSIBLE) site.yml \
		-e validator_ssh_connection=ssh \
		-e validator_ssh_host=$(validator_ssh_host) \
		-e validator_ssh_user=$(or $(validator_ssh_user),hidircan)

# Ansible cannot prompt for sudo; import stays on the Makefile TTY.
define DEPLOY
	mkdir -p $(TF_DIR)/generated
	$(TF) apply -auto-approve $(CONTRACT_TARGETS)
	cd ansible && ../$(ANSIBLE) deploy.yml -e deploy_mode=$(1) --tags build
	sudo bash $(REPO)/scripts/skaffold-k3s-import.sh
	cd ansible && ../$(ANSIBLE) deploy.yml -e deploy_mode=$(1) --tags apply
endef

deploy: $(VENV)
	@if command -v k3s >/dev/null 2>&1 || [ -x /usr/local/bin/k3s ] || [ -x /usr/bin/k3s ]; then \
	  $(MAKE) deploy-k3s; \
	else \
	  echo "k3s yok — bu makinede cluster yok. Docker Compose ile ayağa kalkıyor."; \
	  echo "Cluster istersen: sudo pacman -S k3s && sudo k3s server"; \
	  $(MAKE) up-app; \
	fi

deploy-k3s: $(VENV)
	$(call DEPLOY,cpu)

up-app:
	docker compose --env-file $(REPO)/.env -f $(REPO)/compose.yaml up -d --build

down-app:
	docker compose --env-file $(REPO)/.env -f $(REPO)/compose.yaml down

deploy-gpu: $(VENV)
	$(call DEPLOY,gpu)

deploy-split: $(VENV)
	$(call DEPLOY,split)

forward:
	kubectl -n sadecekedi port-forward svc/go-backend-service 8080:8080
