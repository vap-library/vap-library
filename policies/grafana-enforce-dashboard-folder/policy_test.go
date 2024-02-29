package httproute_enforce_hostnames

import (
	"context"
	"fmt"
	"log"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
	"vap-library/testutils"

	"sigs.k8s.io/e2e-framework/pkg/env"
)

var dashboardCMYAML string = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-dashboard
  namespace: %s
  labels:
    grafana_dashboard: "1"
  annotations:
    grafana_folder: %s
data:
  test: "test"
`

var dashboardCMWithoutAnnotationYAML string = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-dashboard
  namespace: %s
  labels:
    grafana_dashboard: "1"
data:
  test: "test"
`

var normalCMYAML string = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-dashboard
  namespace: %s
data:
  test: "test"
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/grafana-enforce-dashboard-folder": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	os.Exit(testEnv.Run(m))
}

func TestVAPGrafanaEnforceDashboardFolderValid(t *testing.T) {

	f := features.New("Dashboard is accepted").
		Assess("A valid dashboard ConfigMap is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dashboardCMYAML, namespace, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestVAPGrafanaEnforceDashboardFolderNormalCM(t *testing.T) {

	f := features.New("Normal CM is accepted").
		Assess("A non-dashboard CM is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(normalCMYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestVAPGrafanaEnforceDashboardFolderInvalidDashboardCM(t *testing.T) {

	f := features.New("Normal CM is accepted").
		Assess("A non-dashboard CM is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dashboardCMWithoutAnnotationYAML, namespace))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
