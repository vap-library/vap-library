package kustomization_fields

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
kind: VAPLibKustomizationFieldsParam
metadata:
  name: kustomization-fields.vap-library.com
  namespace: %s
spec:
  targetNamespace: app
  serviceAccountName: deployer
`

var kustomizationFullYAML string = `
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: test-kustomization
  namespace: %s
spec:
  interval: 5m
  prune: true
  sourceRef:
    kind: GitRepository
    name: test-repo
  interval: 10m
  targetNamespace: app
  serviceAccountName: deployer
`

var kustomizationWrongSAYAML string = `
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: test-kustomization-wrong-sa
  namespace: %s
spec:
  interval: 5m
  prune: true
  sourceRef:
    kind: GitRepository
    name: test-repo
  interval: 10m
  targetNamespace: app
  serviceAccountName: wrong
`

var testParameterSingleYAML string = `
apiVersion: vap-library.com/v1beta1
kind: VAPLibKustomizationFieldsParam
metadata:
  name: kustomization-fields.vap-library.com
  namespace: %s
spec:
  serviceAccountName: deployer
`

var kustomizationSingleYAML string = `
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: test-kustomization-single-field
  namespace: %s
spec:
  interval: 5m
  prune: true
  sourceRef:
    kind: GitRepository
    name: test-repo
  interval: 10m
  serviceAccountName: deployer
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/kustomization-fields": "deny"}
	var extraResourcesFromDir = map[string]string{"../../vendoring/flux-kustomize-controller/": "*.yaml"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, extraResourcesFromDir)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	// wait for the cluster to be ready
	time.Sleep(3 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestKustomizationNoParam(t *testing.T) {

	f := features.New("Kustomization without parameter").
		Assess("A Kustomization without VAP parameter is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should be rejected!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(kustomizationSingleYAML, namespace))
			if err == nil {
				t.Fatal("A Kustomization without VAP parameter was accepted")
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestKustomizationFull(t *testing.T) {

	f := features.New("Kustomization with full parameter").
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
		Assess("A valid Kustomization with all fields is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(kustomizationFullYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("A Kustomization with missing fields is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(kustomizationSingleYAML, namespace))
			// should not be nil
			if err == nil {
				t.Fatal("A Kustomization with missing fields got accepted")
			}

			// if nil then check if the error message is as expected
			if !strings.HasSuffix(err.Error(), "spec.targetNamespace must be set to the namespace specified in the Validating Admission Policy parameter") {
				t.Fatalf("Kustomization was rejected but with the following, unexpected error message: %s", err.Error())
			}

			return ctx
		}).
		Assess("A Kustomization with wrong fields is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(kustomizationWrongSAYAML, namespace))
			// should not be nil
			if err == nil {
				t.Fatal("A Kustomization with wrong field got accepted")
			}

			// if nil then check if the error message is as expected
			if !strings.HasSuffix(err.Error(), fmt.Sprintf("spec.serviceAccountName must be set to %s. It is: %s", "deployer", "wrong")) {
				t.Fatalf("Kustomization was rejected but with the following, unexpected error message: %s", err.Error())
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestKustomizationSingle(t *testing.T) {

	f := features.New("Kustomization with single parameter").
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
		Assess("A valid Kustomization with single field is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(kustomizationSingleYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
