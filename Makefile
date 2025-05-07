ROOT="github.com/ripta/kubectl-plugins"

CRD_VERSION="v1alpha1"
CRD_NAME="r8y"

KUBERNETES_VERSION=1.33.0
LIBRARY_VERSION=0.33.0

build:
	go build -v -o bin/kubectl-dynaward ./cmd/kubectl-dynaward
	go build -v -o bin/kubectl-show ./cmd/kubectl-show
	go build -v -o bin/kubectl-ssh ./cmd/kubectl-ssh

hyper:
	go build -v -o bin/kp ./hypercmd/kp

update: update-deps update-codegen

update-codegen:
	# go install k8s.io/code-generator/cmd/defaulter-gen@v${LIBRARY_VERSION}
	go install k8s.io/code-generator/cmd/deepcopy-gen@v${LIBRARY_VERSION}
	deepcopy-gen --output-file zz_generated.deepcopy.go --go-header-file /dev/null ./pkg/apis/r8y/v1alpha1

update-deps:
	go get k8s.io/client-go@v$(LIBRARY_VERSION)
	go get ./...
	go mod tidy
