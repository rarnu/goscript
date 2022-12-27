package goscript

import (
	c0 "context"
	"flag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

// TODO: k8s operations

var clientset *kubernetes.Clientset

func init() {
	var kubeConfig *string
	home := homedir.HomeDir()
	kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	// flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err == nil {
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			clientset = nil
		}
	}
}

func (r *Runtime) builtinK8s_allPods(call FunctionCall) Value {
	_ret0 := _null
	if clientset != nil {
		_ns := call.Argument(0).toString().String()
		if _ns == "undefined" {
			_ns = ""
		}
		pods, err := clientset.CoreV1().Pods(_ns).List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			var ms []map[string]any
			for _, pod := range pods.Items {
				m := map[string]any{}
				m["name"] = pod.Name
				m["status"] = pod.Status.Phase
				m["namespace"] = pod.Namespace
				m["labels"] = pod.Labels
				ms = append(ms, m)
			}
			_ret0 = r.ToValue(ms)
		}
	}
	return _ret0
}

func (r *Runtime) builtinK8s_podExists(call FunctionCall) Value {
	_ret0 := false
	if clientset != nil {
		_nm := call.Argument(0).toString().String()
		_ns := call.Argument(1).toString().String()
		if _ns == "undefined" {
			_ns = ""
		}
		pods, err := clientset.CoreV1().Pods(_ns).List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, pod := range pods.Items {
				if pod.Name == _nm {
					_ret0 = true
					break
				}
			}
		}
	}
	return r.toBoolean(_ret0)
}

func (r *Runtime) builtinK8s_podStatus(call FunctionCall) Value {
	_ret0 := ""
	if clientset != nil {
		_nm := call.Argument(0).toString().String()
		_ns := call.Argument(1).toString().String()
		if _ns == "undefined" {
			_ns = ""
		}
		pods, err := clientset.CoreV1().Pods(_ns).List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, pod := range pods.Items {
				if pod.Name == _nm {
					_ret0 = string(pod.Status.Phase)
					break
				}
			}
		}
	}
	return r.ToValue(_ret0)
}

func (r *Runtime) builtinK8s_allServices(call FunctionCall) Value {
	_ret0 := _null
	if clientset != nil {
		_ns := call.Argument(0).toString().String()
		if _ns == "undefined" {
			_ns = ""
		}
		svcs, err := clientset.CoreV1().Services(_ns).List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			var ms []map[string]any
			for _, svc := range svcs.Items {
				m := map[string]any{}
				m["name"] = svc.Name
				m["namespace"] = svc.Namespace
				m["type"] = svc.Spec.Type
				m["labels"] = svc.Labels
				ms = append(ms, m)
			}
			_ret0 = r.ToValue(ms)
		}
	}
	return _ret0
}

func (r *Runtime) builtinK8s_serviceExists(call FunctionCall) Value {
	_ret0 := false
	if clientset != nil {
		_nm := call.Argument(0).toString().String()
		_ns := call.Argument(1).toString().String()
		if _ns == "undefined" {
			_ns = ""
		}
		svcs, err := clientset.CoreV1().Services(_ns).List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, svc := range svcs.Items {
				if svc.Name == _nm {
					_ret0 = true
					break
				}
			}
		}
	}
	return r.toBoolean(_ret0)
}

func (r *Runtime) builtinK8s_allNodes(call FunctionCall) Value {
	_ret0 := _null
	if clientset != nil {
		nodes, err := clientset.CoreV1().Nodes().List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			var ms []map[string]any
			for _, node := range nodes.Items {
				m := map[string]any{}
				m["name"] = node.Name
				m["namespace"] = node.Namespace
				m["status"] = node.Status.Phase
				m["labels"] = node.Labels
				ms = append(ms, m)
			}
			_ret0 = r.ToValue(ms)
		}
	}
	return _ret0
}

func (r *Runtime) builtinK8s_nodeExists(call FunctionCall) Value {
	_ret0 := false
	if clientset != nil {
		_nm := call.Argument(0).toString().String()
		nodes, err := clientset.CoreV1().Nodes().List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, node := range nodes.Items {
				if node.Name == _nm {
					_ret0 = true
					break
				}
			}
		}
	}
	return r.toBoolean(_ret0)
}

func (r *Runtime) builtinK8s_nodeStatus(call FunctionCall) Value {
	_ret0 := ""
	if clientset != nil {
		_nm := call.Argument(0).toString().String()
		nodes, err := clientset.CoreV1().Nodes().List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, node := range nodes.Items {
				if node.Name == _nm {
					_ret0 = string(node.Status.Phase)
					break
				}
			}
		}
	}
	return r.ToValue(_ret0)
}

func (r *Runtime) builtinK8s_allNamespaces(call FunctionCall) Value {
	_ret0 := _null
	if clientset != nil {
		nss, err := clientset.CoreV1().Namespaces().List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			var ms []map[string]any
			for _, ns := range nss.Items {
				m := map[string]any{}
				m["name"] = ns.Name
				m["namespace"] = ns.Namespace
				m["status"] = ns.Status.Phase
				m["labels"] = ns.Labels
				ms = append(ms, m)
			}
			_ret0 = r.ToValue(ms)
		}
	}
	return _ret0
}

func (r *Runtime) builtinK8s_namespaceExists(call FunctionCall) Value {
	_ret0 := false
	if clientset != nil {
		_nm := call.Argument(0).toString().String()
		nss, err := clientset.CoreV1().Namespaces().List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, ns := range nss.Items {
				if ns.Name == _nm {
					_ret0 = true
					break
				}
			}
		}
	}
	return r.toBoolean(_ret0)
}

func (r *Runtime) builtinK8s_namespaceStatus(call FunctionCall) Value {
	_ret0 := ""
	if clientset != nil {
		_nm := call.Argument(0).toString().String()
		nss, err := clientset.CoreV1().Namespaces().List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, ns := range nss.Items {
				if ns.Namespace == _nm {
					_ret0 = string(ns.Status.Phase)
					break
				}
			}
		}
	}
	return r.ToValue(_ret0)
}

func (r *Runtime) builtinK8s_allEndpoints(call FunctionCall) Value {
	_ret0 := _null
	if clientset != nil {
		_ns := call.Argument(0).toString().String()
		if _ns == "undefined" {
			_ns = ""
		}
		eps, err := clientset.CoreV1().Endpoints(_ns).List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			var ms []map[string]any
			for _, ep := range eps.Items {
				m := map[string]any{}
				m["name"] = ep.Name
				m["namespace"] = ep.Namespace
				m["labels"] = ep.Labels
				ms = append(ms, m)
			}
			_ret0 = r.ToValue(ms)
		}
	}
	return _ret0
}

func (r *Runtime) builtinK8s_endpointExists(call FunctionCall) Value {
	_ret0 := false
	if clientset != nil {
		_nm := call.Argument(0).toString().String()
		_ns := call.Argument(1).toString().String()
		if _ns == "undefined" {
			_ns = ""
		}
		eps, err := clientset.CoreV1().Endpoints(_ns).List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, ep := range eps.Items {
				if ep.Name == _nm {
					_ret0 = true
					break
				}
			}
		}
	}
	return r.toBoolean(_ret0)
}

func (r *Runtime) builtinK8s_allEvents(call FunctionCall) Value {
	_ret0 := _null
	if clientset != nil {
		_ns := call.Argument(0).toString().String()
		if _ns == "undefined" {
			_ns = ""
		}
		evs, err := clientset.CoreV1().Events(_ns).List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			var ms []map[string]any
			for _, ev := range evs.Items {
				m := map[string]any{}
				m["name"] = ev.Name
				m["namespace"] = ev.Namespace
				m["type"] = ev.Type
				m["labels"] = ev.Labels
				ms = append(ms, m)
			}
			_ret0 = r.ToValue(ms)
		}
	}
	return _ret0
}

func (r *Runtime) builtinK8s_eventExists(call FunctionCall) Value {
	_ret0 := false
	if clientset != nil {
		_nm := call.Argument(0).toString().String()
		_ns := call.Argument(1).toString().String()
		if _ns == "undefined" {
			_ns = ""
		}
		evs, err := clientset.CoreV1().Events(_ns).List(c0.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, ev := range evs.Items {
				if ev.Name == _nm {
					_ret0 = true
					break
				}
			}
		}
	}
	return r.toBoolean(_ret0)
}

func (r *Runtime) initK8s() {
	K8S := r.newBaseObject(r.global.ObjectPrototype, "Kubernetes")
	K8S._putProp("allPods", r.newNativeFunc(r.builtinK8s_allPods, nil, "allPods", nil, 1), true, false, true)
	K8S._putProp("podExists", r.newNativeFunc(r.builtinK8s_podExists, nil, "podExists", nil, 2), true, false, true)
	K8S._putProp("podStatus", r.newNativeFunc(r.builtinK8s_podStatus, nil, "podStatus", nil, 2), true, false, true)
	K8S._putProp("allServices", r.newNativeFunc(r.builtinK8s_allServices, nil, "allServices", nil, 1), true, false, true)
	K8S._putProp("serviceExists", r.newNativeFunc(r.builtinK8s_serviceExists, nil, "serviceExists", nil, 2), true, false, true)
	K8S._putProp("allNodes", r.newNativeFunc(r.builtinK8s_allNodes, nil, "allNodes", nil, 0), true, false, true)
	K8S._putProp("nodeExists", r.newNativeFunc(r.builtinK8s_nodeExists, nil, "nodeExists", nil, 1), true, false, true)
	K8S._putProp("nodeStatus", r.newNativeFunc(r.builtinK8s_nodeStatus, nil, "nodeStatus", nil, 1), true, false, true)
	K8S._putProp("allNamespaces", r.newNativeFunc(r.builtinK8s_allNamespaces, nil, "allNamespaces", nil, 0), true, false, true)
	K8S._putProp("namespaceExists", r.newNativeFunc(r.builtinK8s_namespaceExists, nil, "namespaceExists", nil, 1), true, false, true)
	K8S._putProp("namespaceStatus", r.newNativeFunc(r.builtinK8s_namespaceStatus, nil, "namespaceStatus", nil, 1), true, false, true)
	K8S._putProp("allEndpoints", r.newNativeFunc(r.builtinK8s_allEndpoints, nil, "allEndpoints", nil, 1), true, false, true)
	K8S._putProp("endpointExists", r.newNativeFunc(r.builtinK8s_endpointExists, nil, "endpointExists", nil, 2), true, false, true)
	K8S._putProp("allEvents", r.newNativeFunc(r.builtinK8s_allEvents, nil, "allEvents", nil, 1), true, false, true)
	K8S._putProp("eventExists", r.newNativeFunc(r.builtinK8s_eventExists, nil, "eventExists", nil, 2), true, false, true)
	r.addToGlobal("Kubernetes", K8S.val)
}
