CONTROLLER_VERSION=v0.1.0

release: deploy
build: codegen dockerbuild
dockerbuild: cli controller minecraftserver
clean: teardown deploy

codegen:
	@echo "Generating code..."
	#@controller-gen object:headerFile=./hack/boilerplate.go.txt paths="./..."
	@./hack/update-codegen.sh

publish_images:
	@echo "Publishing images to minikube..."
	minikube image load kubecraft/kubecraft-controller:${CONTROLLER_VERSION}
	minikube image load kubecraft/kubecraft-cli:${CONTROLLER_VERSION}
	minikube image load kubecraft/minecraft-server:${CONTROLLER_VERSION}

gencrd:
	@echo "Generating CRD..."
	@controller-gen crd:trivialVersions=true paths="./..."

crd:
	@echo "Applying CRD..."
	@kubectl apply -f artifacts/examples/crd-minecraftserver.yaml

controller:
	@echo "Building Controller Container Image..."
	@docker build -t kubecraft/kubecraft-controller:${CONTROLLER_VERSION} .
	minikube image rm kubecraft/kubecraft-controller:${CONTROLLER_VERSION}
	minikube image load kubecraft/kubecraft-controller:${CONTROLLER_VERSION}

cli:
	@echo "Building CLI Container Image..."
	@docker build -t kubecraft/kubecraft-cli:${CONTROLLER_VERSION} -f Dockerfile.cli .
	minikube image rm kubecraft/kubecraft-cli:${CONTROLLER_VERSION}
	minikube image load kubecraft/kubecraft-cli:${CONTROLLER_VERSION}

minecraftserver:
	@echo "Building Minecraft Server Image..."
	@docker build -t kubecraft/minecraft-server:${CONTROLLER_VERSION} -f Dockerfile.minecraft .
	minikube image rm kubecraft/minecraft-server:${CONTROLLER_VERSION}
	minikube image load kubecraft/minecraft-server:${CONTROLLER_VERSION}

deploy:
	@echo "Deploying Controller..."
	@kubectl apply -f artifacts/examples/crd-minecraftserver.yaml
	@kubectl apply -f kubecraft-deployment.yaml
	@kubectl apply -f artifacts/examples/mincraftserver.yaml

teardown:
	@echo "Tearing down Controller..."
	@kubectl delete -f artifacts/examples/crd-minecraftserver.yaml
	@kubectl delete -f kubecraft-deployment.yaml


