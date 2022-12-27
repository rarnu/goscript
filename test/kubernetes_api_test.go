package test

import (
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/console"
	"github.com/rarnu/goscript/require"
	"testing"
)

func TestK8sPods(t *testing.T) {
	SCRIPT := "let pods = Kubernetes.allPods()\nfor (let i=0; i<pods.length; i++) {\n    console.log(pods[i].name)\n}\n"
	_vm := goscript.New()
	new(require.Registry).Enable(_vm)
	console.Enable(_vm)
	_, _ = _vm.RunString(SCRIPT)
}

func TestK8sSvcs(t *testing.T) {
	SCRIPT := "let svcs = Kubernetes.allServices()\nfor (let i=0; i<svcs.length; i++) {\n    console.log(svcs[i].name)\n}\n"
	_vm := goscript.New()
	new(require.Registry).Enable(_vm)
	console.Enable(_vm)
	_, _ = _vm.RunString(SCRIPT)
}

func TestK8sNodes(t *testing.T) {
	SCRIPT := "let nodes = Kubernetes.allNodes()\nfor (let i=0; i<nodes.length; i++) {\n    console.log(nodes[i].name)\n}\n"
	_vm := goscript.New()
	new(require.Registry).Enable(_vm)
	console.Enable(_vm)
	_, _ = _vm.RunString(SCRIPT)
}

func TestK8sNamespaces(t *testing.T) {
	SCRIPT := "let nss = Kubernetes.allNamespaces()\nfor (let i=0; i<nss.length; i++) {\n    console.log(nss[i].name)\n}\n"
	_vm := goscript.New()
	new(require.Registry).Enable(_vm)
	console.Enable(_vm)
	_, _ = _vm.RunString(SCRIPT)
}

func TestK8sEndpoints(t *testing.T) {
	SCRIPT := "let eps = Kubernetes.allEndpoints()\nfor (let i=0; i<eps.length; i++) {\n    console.log(eps[i].name)\n}\n"
	_vm := goscript.New()
	new(require.Registry).Enable(_vm)
	console.Enable(_vm)
	_, _ = _vm.RunString(SCRIPT)
}

func TestK8sEvents(t *testing.T) {
	SCRIPT := "let evs = Kubernetes.allEvents()\nfor (let i=0; i<evs.length; i++) {\n    console.log(evs[i].name)\n}\n"
	_vm := goscript.New()
	new(require.Registry).Enable(_vm)
	console.Enable(_vm)
	_, _ = _vm.RunString(SCRIPT)
}
