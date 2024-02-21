package main

import (
	"os"
	"strings"
	"testing"
	"vap-library/testutils"

	"sigs.k8s.io/e2e-framework/pkg/env"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()
	testutils.CheckPrerequisites()
	testutils.CreateKindCluster()
	os.Exit(testenv.Run(m))
}

func TestVapGrafanaEnforceDashboardFolder(t *testing.T) {
	testutils.CreateFromFile("policy.yaml", t)
	testutils.CreateFromFile("binding.yaml", t)
	testutils.DeleteNamespace("grafana-enforce-dashboard-folder-vap-library-test", t)
	testutils.CreateFromFile("namespace.yaml", t)

	t.Run("dashboard with folder corresponding to namespace should be allowed", func(t *testing.T) {
		testutils.CreationShouldSucceed(t, testutils.Dedent(`
			apiVersion: v1
			kind: ConfigMap
			metadata:
				name: dashboard-in-correct-folder
				namespace: grafana-enforce-dashboard-folder-vap-library-test
				labels:
					grafana_dashboard: "1"
				annotations:
					grafana_folder: grafana-enforce-dashboard-folder-vap-library-test
			data:
				test: "test"`))
	})

	t.Run("dashboard with folder different from namespace should be denied", func(t *testing.T) {
		errorMessage := testutils.CreationShouldFail(t, testutils.Dedent(`
			apiVersion: v1
			kind: ConfigMap
			metadata:
				name: dashboard-in-wrong-folder
				namespace: grafana-enforce-dashboard-folder-vap-library-test
				labels:
					grafana_dashboard: "1"
				annotations:
					grafana_folder: some-other-folder
			data:
				test: "test"`))

		if !strings.HasSuffix(errorMessage, "metadata.annotations.grafana_folder must be set to the namespace of the ConfigMap/Secret\n") {
			t.Errorf("Unexpected error message: %s", errorMessage)
		}
	})

	t.Run("dashboard with no folder specified should be denied", func(t *testing.T) {
		errorMessage := testutils.CreationShouldFail(t, testutils.Dedent(`
			apiVersion: v1
			kind: ConfigMap
			metadata:
				name: grafana-enforce-dashboard-folder-vap-library-test
				namespace: grafana-enforce-dashboard-folder-vap-library-test
				labels:
					grafana_dashboard: "1"
			data:
				test: "test"`))

		if !strings.HasSuffix(errorMessage, "metadata.annotations.grafana_folder must be set to the namespace of the ConfigMap/Secret\n") {
			t.Errorf("Unexpected error message: %s", errorMessage)
		}
	})
}
