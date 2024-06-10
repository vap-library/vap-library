package helmrelease_fields

import (
	"context"
	"fmt"
	"log"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"strings"
	"testing"
	"time"
	"vap-library/testutils"
)

var testParameterFullYAML string = `
apiVersion: vap-library.com/v1beta1
kind: VAPLibHelmReleaseFieldsParam
metadata:
  name: helmrelease-fields.vap-library.com
  namespace: %s
spec:
  targetNamespace: app
  serviceAccountName: deployer
`

var helmReleaseFullYAML string = `
apiVersion: helm.toolkit.fluxcd.io/v2beta2
kind: HelmRelease
metadata:
  name: test-helmrelease
  namespace: %s
spec:
  chart:
    spec:
      chart: example
      interval: 5m
      sourceRef:
        kind: HelmRepository
        name: test-repo
      version: 0.1.3
  interval: 10m
  targetNamespace: app
  serviceAccountName: deployer
`

var helmReleaseWrongSAYAML string = `
apiVersion: helm.toolkit.fluxcd.io/v2beta2
kind: HelmRelease
metadata:
  name: test-helmrelease-wrong-sa
  namespace: %s
spec:
  chart:
    spec:
      chart: example
      interval: 5m
      sourceRef:
        kind: HelmRepository
        name: test-repo
      version: 0.1.3
  interval: 10m
  targetNamespace: app
  serviceAccountName: wrong
`

var testParameterSingleYAML string = `
apiVersion: vap-library.com/v1beta1
kind: VAPLibHelmReleaseFieldsParam
metadata:
  name: helmrelease-fields.vap-library.com
  namespace: %s
spec:
  serviceAccountName: deployer
`

var helmReleaseSingleYAML string = `
apiVersion: helm.toolkit.fluxcd.io/v2beta2
kind: HelmRelease
metadata:
  name: test-helmrelease-single-field
  namespace: %s
spec:
  chart:
    spec:
      chart: example
      interval: 5m
      sourceRef:
        kind: HelmRepository
        name: test-repo
      version: 0.1.3
  interval: 10m
  serviceAccountName: deployer
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/helmrelease-fields": "deny"}
	var extraResourcesFromDir = map[string]string{"../../vendoring/flux-helm-controller/": "*.yaml"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, extraResourcesFromDir)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	// wait for the cluster to be ready
	time.Sleep(3 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestHelmReleaseNoParam(t *testing.T) {

	f := features.New("HelmRelease without parameter").
		Assess("A HelmRelease without VAP parameter is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should be rejected!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(helmReleaseSingleYAML, namespace))
			if err == nil {
				t.Fatal("A HelmRelease without VAP parameter was accepted")
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestHelmReleaseFull(t *testing.T) {

	f := features.New("HelmRelease with full parameter").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// apply parameter first
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(testParameterFullYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the parameter to be registered properly
			time.Sleep(10 * time.Second)

			return ctx
		}).
		Assess("A valid HelmRelease with all fields is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(helmReleaseFullYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("A HelmRelease with missing fields is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(helmReleaseSingleYAML, namespace))
			// should not be nil
			if err == nil {
				t.Fatal("A HelmRelease with missing fields got accepted")
			}

			// if nil then check if the error message is as expected
			if !strings.HasSuffix(err.Error(), "spec.targetNamespace must be set to the namespace specified in the Validating Admission Policy parameter") {
				t.Fatalf("HelmRelease was rejected but with the following, unexpected error message: %s", err.Error())
			}

			return ctx
		}).
		Assess("A HelmRelease with wrong fields is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(helmReleaseWrongSAYAML, namespace))
			// should not be nil
			if err == nil {
				t.Fatal("A HelmRelease with wrong field got accepted")
			}

			// if nil then check if the error message is as expected
			if !strings.HasSuffix(err.Error(), fmt.Sprintf("spec.serviceAccountName must be set to %s. It is: %s", "deployer", "wrong")) {
				t.Fatalf("HelmRelease was rejected but with the following, unexpected error message: %s", err.Error())
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestHelmReleaseSingle(t *testing.T) {

	f := features.New("HelmRelease with single parameter").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// apply parameter first
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(testParameterSingleYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the parameter to be registered properly
			time.Sleep(10 * time.Second)

			return ctx
		}).
		Assess("A valid HelmRelease with single field is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(helmReleaseSingleYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
