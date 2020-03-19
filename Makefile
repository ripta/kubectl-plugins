CODEGEN="vendor/k8s.io/code-generator"
ROOT="github.com/ripta/kubectl-plugins"

CRD_VERSION="v1alpha1"
CRD_NAME="r8y"

KUBERNETES_VERSION=1.12.2

build:
	go build -v -o bin/kubectl-show ./cmd/kubectl-show
	go build -v -o bin/kubectl-ssh ./cmd/kubectl-ssh

hyper:
	go build -v -o bin/ripta-kubectl-plugins ./hyperbinary

update: update-deps update-codegen

update-codegen:
	[ -d $(CODEGEN) ] || git clone -b v0.17.2 https://github.com/kubernetes/code-generator vendor/k8s.io/code-generator
	cd $(CODEGEN) && go install ./cmd/{defaulter-gen,deepcopy-gen}
	deepcopy-gen --bounding-dirs . --input-dirs ./pkg/apis/r8y/v1alpha1 --output-base . --output-file-base zz_generated.deepcopy --go-header-file /dev/null

update-deps:
	go get -u ./...

