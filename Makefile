CODEGEN="vendor/k8s.io/code-generator"
ROOT="github.com/ripta/kubectl-plugins"

CRD_VERSION="v1alpha1"
CRD_NAME="r8y"

update-codegen:
	$(CODEGEN)/generate-groups.sh deepcopy $(ROOT)/pkg/client $(ROOT)/pkg/apis $(CRD_NAME):$(CRD_VERSION)
