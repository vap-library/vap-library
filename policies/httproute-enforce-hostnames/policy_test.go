package httproute_enforce_hostnames

import (
	"context"
	"fmt"
	"log"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
	"time"
	"vap-library/testutils"
)

var testParameterYAML string = `
apiVersion: vap-library.com/v1beta1
kind: HTTPRouteEnforceHostnamesParam
metadata:
  name: httproute-enforce-hostnames-vap-library-test
  namespace: %s
spec:
  allowedHostnames:
  - test.example.com
  - test2.example.com
`

var validHostnameYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute-example-vap-library-test-test01
  namespace: %s
spec:
  hostnames:
  - test.example.com
  parentRefs:
  - name: dummy
`

var invalidHostnameYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute-example-vap-library-test-test01
  namespace: %s
spec:
  hostnames:
  - notallowed.example.com
  parentRefs:
  - name: dummy
`

var noHostnameYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute-example-vap-library-test-test01
  namespace: %s
spec:
  parentRefs:
  - name: dummy
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/httproute-enforce-hostnames": "deny"}
	var extraResourcesFromDir = map[string]string{"../../vendoring/gateway-api/": "*.yaml"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, extraResourcesFromDir)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestValidHostname(t *testing.T) {

	f := features.New("HTTPRoute is accepted").
		Assess("A valid HTTPRoute is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// apply parameter first
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(testParameterYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the parameter to be registered properly
			time.Sleep(10 * time.Second)

			// this should PASS!
			err = testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(validHostnameYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestInValidHostname(t *testing.T) {

	f := features.New("Invalid HTTPRoute is rejected").
		Assess("A HTTPRoute with invalid hostname is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// apply parameter first
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(testParameterYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the parameter to be registered properly
			time.Sleep(10 * time.Second)

			// this should FAIL!
			err = testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(invalidHostnameYAML, namespace))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestWithoutParameter(t *testing.T) {

	f := features.New("Requests without parameter are rejected").
		Assess("A without the VAP parameter, HTTPRoutes are rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL as we do not have a parameter for VAP!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(validHostnameYAML, namespace))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestNoHostname(t *testing.T) {

	f := features.New("HTTPRoute without hostname is rejected").
		Assess("A HTTPRoute which does not contain any hostname is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// apply parameter first
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(testParameterYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the parameter to be registered properly
			time.Sleep(10 * time.Second)

			// this should FAIL!
			err = testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(noHostnameYAML, namespace))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
