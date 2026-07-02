APP        := tool-agent
VERSION    := latest
NAMESPACE  := xrj-app

# 部署文件在外部仓库 /Users/jiang/jiangrx816/minikube
DEPLOY_DIR := /Users/jiang/jiangrx816/minikube/app
CONFIG_YML := $(DEPLOY_DIR)/config/app.yml

.PHONY: help build load deploy-config deploy redeploy status logs port-forward

help:
	@echo "可用命令:"
	@echo "  make build          构建 docker 镜像 tool-agent:latest"
	@echo "  make load           将镜像导入 minikube"
	@echo "  make deploy-config  根据 $(CONFIG_YML) 创建/更新 ConfigMap"
	@echo "  make deploy         应用 tool-agent.yaml"
	@echo "  make redeploy       重新构建+加载+滚动重启"
	@echo "  make status         查看 Pod 状态"
	@echo "  make logs           跟随查看 tool-agent 日志"
	@echo "  make port-forward   本机 8080 -> Pod 8080"

build:
	docker build -t $(APP):$(VERSION) .

load:
	minikube image load $(APP):$(VERSION)

# 用 --dry-run=client -o yaml | kubectl apply 实现 idempotent 的 create/update
deploy-config:
	kubectl create configmap tool-agent-config \
		-n $(NAMESPACE) \
		--from-file=app.yml=$(CONFIG_YML) \
		--dry-run=client -o yaml | kubectl apply -f -

deploy:
	kubectl apply -f $(DEPLOY_DIR)/tool-agent.yaml
	@echo ""
	@echo "提示: 暴露访问使用  make port-forward"
	@echo "      或直接 NodePort: minikube service tool-agent -n $(NAMESPACE)"

redeploy: build load
	kubectl rollout restart deployment/tool-agent -n $(NAMESPACE)

status:
	kubectl get pods,svc -n $(NAMESPACE)

logs:
	kubectl logs -f deployment/tool-agent -n $(NAMESPACE)

port-forward:
	kubectl port-forward svc/tool-agent -n $(NAMESPACE) 8080:8080
