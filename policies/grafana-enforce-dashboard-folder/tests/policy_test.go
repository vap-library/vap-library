package main

import (
	"os"
	"strings"
	"testing"
	"vap-library/testutils"

	"github.com/lithammer/dedent"
	"sigs.k8s.io/e2e-framework/pkg/env"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()
	// kindClusterName := envconf.RandomName("vap-library-", 16)
	// testenv.Setup(
	// 	envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
	// )

	// testenv.Finish(
	// 	envfuncs.DestroyCluster(kindClusterName),
	// )

	os.Exit(testenv.Run(m))
}

func TestVapGrafanaEnforceDashboardFolder(t *testing.T) {
	testutils.CreateFromFile("../policy.yaml", t)
	testutils.CreateFromFile("binding.yaml", t)
	testutils.RecreateNamespace("my-namespace", t)

	t.Run("dashboard with folder corresponding to namespace should be allowed", func(t *testing.T) {
		testutils.CreationShouldSucceed(t, dedent.Dedent(`
			apiVersion: v1
			kind: ConfigMap
			metadata:
			  name: dashboard-in-correct-folder
			  namespace: my-namespace
			  labels:
			    grafana_dashboard: "1"
			  annotations:
			    grafana_folder: my-namespace
			data:
			  test: "test"`))
	})

	t.Run("dashboard with folder different from namespace should be denied", func(t *testing.T) {
		errorMessage := testutils.CreationShouldFail(t, dedent.Dedent(`
			apiVersion: v1
			kind: ConfigMap
			metadata:
			  name: dashboard-in-wrong-folder
			  namespace: my-namespace
			  labels:
			    grafana_dashboard: "1"
			  annotations:
			    grafana_folder: "some-other-folder"
			data:
			  test: "test"`))

		if !strings.HasSuffix(errorMessage, "metadata.annotations.grafana_folder must be set to the namespace of the ConfigMap/Secret\n") {
			t.Errorf("Unexpected error message: %s", errorMessage)
		}
	})
}
