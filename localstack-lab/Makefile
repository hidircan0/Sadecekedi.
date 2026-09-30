# Sadecekedi LocalStack & k3s Otomasyon Makefile
# Bu dosya tüm sistemi tek komutla temiz bir şekilde açıp kapatmayı sağlar.

ENDPOINT := http://127.0.0.1:4566
TF_DIR := ansible/terraform
TF := terraform -chdir=$(TF_DIR)
REPO := $(abspath $(CURDIR)/..)
VENV := .venv
ANSIBLE := $(VENV)/bin/ansible-playbook
export ANSIBLE_HOME := $(CURDIR)/.ansible

.DEFAULT_GOAL := help
.PHONY: help up down $(VENV)

help:
	@echo "Kullanılabilir Komutlar:"
	@echo "  make up      Sıfırdan tüm sistemi ayağa kaldırır (LocalStack + Terraform + k3s)"
	@echo "  make down    Tüm sistemi temiz bir şekilde kapatır (Port tünelleri + Docker + k3s namespace)"

$(VENV):
	@python3 -m venv $(VENV)
	@$(VENV)/bin/pip install -U pip ansible-core >/dev/null

up: $(VENV)
	@echo "=== [1/7] K8s Secret Tanımları Hazırlanıyor ==="
	@test -f $(REPO)/k8s/01-config-secrets.yaml || $(REPO)/scripts/render-k8s-secrets.sh
	
	@echo "=== [2/7] LocalStack Başlatılıyor ==="
	@docker compose up -d
	@echo "LocalStack sağlık kontrolü bekleniyor..."
	@i=0; until curl -sf $(ENDPOINT)/_localstack/health >/dev/null; do \
		i=$$((i+1)); if [ $$i -gt 90 ]; then echo "HATA: LocalStack 4566 portu açılmadı!"; exit 1; fi; sleep 2; \
	done
	@echo "LocalStack Hazır!"

	@echo "=== [3/7] Terraform Altyapısı Dağıtılıyor ==="
	@mkdir -p $(TF_DIR)/generated
	@export TF_VAR_ssh_public_key=$$(grep SSH_PUBLIC_KEY ../.env | cut -d '=' -f2 | tr -d '"'); \
	$(TF) init -input=false && $(TF) apply -auto-approve

	@echo "=== [4/7] Docker İmajları Derleniyor (Ansible Build) ==="
	@cd ansible && ../$(ANSIBLE) deploy.yml -e deploy_mode=cpu --tags build

	@echo "=== [5/7] İmajlar k3s Kümesine Aktarılıyor (Import) ==="
	@sudo bash $(REPO)/scripts/skaffold-k3s-import.sh

	@echo "=== [6/7] Kubernetes Manifestleri Uygulanıyor (Ansible Apply) ==="
	@cd ansible && ../$(ANSIBLE) deploy.yml -e deploy_mode=cpu --tags apply

	@echo "=== [7/7] Yerel Port Yönlendirmeleri Başlatılıyor ==="
	@sudo pkill -f "kubectl port-forward" || true
	@nohup sudo k3s kubectl port-forward -n sadecekedi svc/go-backend-service 8080:8080 >/dev/null 2>&1 &
	@nohup sudo k3s kubectl port-forward -n sadecekedi svc/grafana-service 3000:3000 >/dev/null 2>&1 &
	@nohup sudo k3s kubectl port-forward -n kube-system svc/traefik 9080:80 >/dev/null 2>&1 &
	
	@echo ""
	@echo "======================================================="
	@echo " KURULUM BAŞARIYLA TAMAMLANDI! 🚀"
	@echo "======================================================="
	@echo " - Go Backend API : http://localhost:8080"
	@echo " - Grafana Panel  : http://localhost:3000 (User: hidir / Şifre: bosalpbora)"
	@echo " - Traefik Panel  : http://localhost:9080/dashboard/"
	@echo "======================================================="

down:
	@echo "=== [1/3] Port Yönlendirme Süreçleri Kapatılıyor ==="
	-@sudo pkill -f "kubectl port-forward" 2>/dev/null || true

	@echo "=== [2/3] LocalStack Konteynırı Durduruluyor ==="
	@docker compose down

	@echo "=== [3/3] Kubernetes Namespace Temizleniyor ==="
	-@sudo k3s kubectl delete namespace sadecekedi --ignore-not-found
	@echo "Sistem temizlendi! 🧹"
