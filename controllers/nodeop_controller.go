/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"os/exec"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodesv1alpha1 "github.com/jike-inc/node-operator/api/v1alpha1"
)

// NodeOPReconciler reconciles a NodeOP object
type NodeOPReconciler struct {
	HostName string
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nodes.jike.com,resources=nodeops,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodes.jike.com,resources=nodeops/status,verbs=get;update;patch

func (r *NodeOPReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	l := r.Log.WithValues("memcached", req.NamespacedName)
	nodeOP := &nodesv1alpha1.NodeOP{}
	err := r.Get(ctx, req.NamespacedName, nodeOP)
	if err != nil {
		l.Error(err, "get cr error")
		return ctrl.Result{}, nil
	}
	if nodeOP.Name != r.HostName {
		return ctrl.Result{}, nil
	}

	l.Info(fmt.Sprintf("%s,%v", nodeOP.Spec.Command, nodeOP.Spec.Args))
	out, err := r.RunCommand(nodeOP.Spec.Command, nodeOP.Spec.Args)
	if err != nil {
		l.Error(err, "Exec command error")
		return ctrl.Result{}, nil
	}
	fmt.Printf("combined out:\n%s\n", out)
	return ctrl.Result{}, nil
}

func (r *NodeOPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nodesv1alpha1.NodeOP{}).
		Complete(r)
}

//RunCommand exec command
func (r *NodeOPReconciler) RunCommand(cmd string, args []string) (string, error) {
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
