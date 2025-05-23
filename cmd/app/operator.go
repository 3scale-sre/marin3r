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

package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/3scale-sre/basereconciler/reconciler"
	"github.com/spf13/cobra"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	marin3rv1alpha1 "github.com/3scale-sre/marin3r/api/marin3r/v1alpha1"
	operatorv1alpha1 "github.com/3scale-sre/marin3r/api/operator.marin3r/v1alpha1"
	operatorcontroller "github.com/3scale-sre/marin3r/internal/controller/operator.marin3r"
	// +kubebuilder:scaffold:imports
)

var (
	leaderElect    bool
	operatorScheme = apimachineryruntime.NewScheme()
)

var (
	// Operator subcommand
	operatorCmd = &cobra.Command{
		Use:   "operator",
		Short: "Run the operator",
		Run:   runOperator,
	}
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(operatorScheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(operatorScheme))
	utilruntime.Must(marin3rv1alpha1.AddToScheme(operatorScheme))
	// +kubebuilder:scaffold:scheme

	rootCmd.AddCommand(operatorCmd)

	// Operator flags
	operatorCmd.Flags().BoolVar(&leaderElect, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
}

func runOperator(cmd *cobra.Command, args []string) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(debug)))
	printVersion()

	cfg := ctrl.GetConfigOrDie()

	watchNamespace, err := getWatchNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all Namespaces")
	}

	options := ctrl.Options{
		Scheme: operatorScheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress:     probeAddr,
		LeaderElection:             leaderElect,
		LeaderElectionID:           "2cfbe7d6.operator.marin3r.3scale.net",
		LeaderElectionResourceLock: "leases",
	}

	if strings.Contains(watchNamespace, ",") {
		setupLog.Info(fmt.Sprintf("manager in MultiNamespaced mode will be watching namespaces %q", watchNamespace))

		options.Cache = cache.Options{DefaultNamespaces: map[string]cache.Config{}}
		for _, ns := range strings.Split(watchNamespace, ",") {
			options.Cache.DefaultNamespaces[ns] = cache.Config{}
		}
	} else if watchNamespace != "" {
		setupLog.Info(fmt.Sprintf("manager in Namespaced mode will be watching namespace %q", watchNamespace))
		options.Cache = cache.Options{DefaultNamespaces: map[string]cache.Config{
			watchNamespace: {},
		}}
	} else {
		setupLog.Info("manager in Cluster scope mode will be watching all namespaces")
	}

	mgr, err := ctrl.NewManager(cfg, options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := (&operatorcontroller.DiscoveryServiceReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("discoveryservice")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "discoveryservice")
		os.Exit(1)
	}

	if err := (&operatorcontroller.DiscoveryServiceCertificateReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("discoveryservicecertificate")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "discoveryservicecertificate")
		os.Exit(1)
	}

	if err = (&operatorcontroller.EnvoyDeploymentReconciler{
		Reconciler: reconciler.NewFromManager(mgr).
			WithLogger(ctrl.Log.WithName("controllers").WithName("envoydeployment")),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "EnvoyDeployment")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting the Operator.")

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "controller manager exited non-zero")
		os.Exit(1)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}

	return ns, nil
}
