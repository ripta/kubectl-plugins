module github.com/ripta/kubectl-plugins

go 1.13

require (
	github.com/itchyny/gojq v0.10.1
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/pkg/errors v0.9.1
	github.com/ripta/hypercmd v0.0.0-20200525050207-eb5806825d0c
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.17.5
	k8s.io/apimachinery v0.26.3
	k8s.io/cli-runtime v0.17.5
	k8s.io/client-go v0.17.5
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.17.5
)
