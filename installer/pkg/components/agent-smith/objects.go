package agentsmith

import "github.com/gitpod-io/gitpod/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	configmap,
	daemonset,
	networkpolicy,
	role,
	rolebinding,
	common.DefaultServiceAccount(Component),
)
