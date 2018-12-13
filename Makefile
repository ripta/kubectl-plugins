CODEGEN="vendor/k8s.io/code-generator"
ROOT="github.com/ripta/kubectl-plugins"

CRD_VERSION="v1alpha1"
CRD_NAME="r8y"

KUBERNETES_VERSION=1.12.3

build:
	go build -v -o bin/kubectl-show github.com/ripta/kubectl-plugins/cmd/kubectl-show
	go build -v -o bin/kubectl-ssh github.com/ripta/kubectl-plugins/cmd/kubectl-ssh

ensure:
	dep ensure -v

refresh-pkg:
	echo "# DO NOT MODIFY THIS FILE" > Gopkg.toml
	echo "# Change Gopkg.header.toml and then run 'make refresh-pkg' instead" >> Gopkg.toml
	echo "# Generated $$(date)" >> Gopkg.toml
	echo >> Gopkg.toml
	cat Gopkg.header.toml | sed -e "s#KUBEVERSION#$(KUBERNETES_VERSION)#g" >> Gopkg.toml
	echo >> Gopkg.toml
	echo "# Godeps.json extracted from client-go" >> Gopkg.toml
	curl -L https://raw.githubusercontent.com/kubernetes/client-go/kubernetes-$(KUBERNETES_VERSION)/Godeps/Godeps.json \
		| hack/normalize.py >> Gopkg.toml

update-codegen:
	$(CODEGEN)/generate-groups.sh deepcopy $(ROOT)/pkg/client $(ROOT)/pkg/apis $(CRD_NAME):$(CRD_VERSION)
