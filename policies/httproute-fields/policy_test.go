package httproute_hostnames

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

// Variables for hostname tests
var testParameterHostnameYAML string = `
apiVersion: vap-library.com/v1beta1
kind: VAPLibHTTPRouteFieldsParam
metadata:
  name: httproute-fields.vap-library.com
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
  name: test-httproute
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
  name: test-httproute-invalid-hostname
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
  name: test-httproute-no-hostname
  namespace: %s
spec:
  parentRefs:
  - name: dummy
`

// Variables for parentRef tests
var testParameterParentRefYAML string = `
apiVersion: vap-library.com/v1beta1
kind: VAPLibHTTPRouteFieldsParam
metadata:
  name: httproute-fields.vap-library.com
  namespace: %s
spec:
  allowedParentRefs:
  - name: name-only-gateway
  - name: with-namespace-gateway
    namespace: gateway-namespace
`

// PASS: name is correct, no namespace defined
var validNameGatewayYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-httproute
  namespace: %s
spec:
  parentRefs:
  - name: name-only-gateway
`

// PASS: name and namespace are correct
var validNameAndNamespaceGatewayYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-httproute
  namespace: %s
spec:
  parentRefs:
  - name: with-namespace-gateway
    namespace: gateway-namespace
`

// PASS: name is right, and we ignore everything what is not defined in the parameter
var wrongNamespaceGatewayYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-httproute
  namespace: %s
spec:
  parentRefs:
  - name: name-only-gateway
    namespace: should-be-ignored
`

// PASS: multiple parentRefs and all good
var validMultiGatewayYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-httproute
  namespace: %s
spec:
  parentRefs:
  - name: name-only-gateway
  - name: with-namespace-gateway
    namespace: gateway-namespace
`

// FAIL: name is wrong
var wrongNameGatewayYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-httproute
  namespace: %s
spec:
  parentRefs:
  - name: dummy
`

// FAIL: name is right but namespace is taken from the other allowed parent
var wrongMixedGatewayYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-httproute
  namespace: %s
spec:
  parentRefs:
  - name: name-only-gateway
    namespace: gateway-namespace
`

// FAIL: one item is right the other is wrong
var wrongMultiGatewayYAML string = `
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-httproute
  namespace: %s
spec:
  parentRefs:
  - name: name-only-gateway
  - name: wrong-gateway
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/httproute-fields": "deny"}
	var extraResourcesFromDir = map[string]string{"../../vendoring/gateway-api/": "*.yaml"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, extraResourcesFromDir)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestWithParameter(t *testing.T) {

	f := features.New("HTTPRoute with parameter").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// apply parameter first
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(testParameterHostnameYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the parameter to be registered properly
			time.Sleep(10 * time.Second)

			return ctx
		}).
		Assess("A valid HTTPRoute is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(validHostnameYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("An HTTPRoute with invalid hostname is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(invalidHostnameYAML, namespace))
			if err == nil {
				t.Fatal("An HTTPRoute with invalid hostname was accepted")
			}

			return ctx
		}).
		Assess("A HTTPRoute which does not contain any hostname is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(noHostnameYAML, namespace))
			if err == nil {
				t.Fatal("An HTTPRoute without hostname was accepted")
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestWithoutParameter(t *testing.T) {

	f := features.New("HTTPRoute without VAP parameter").
		Assess("Without the VAP parameter, HTTPRoutes are rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL as we do not have a parameter for VAP!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(validHostnameYAML, namespace))
			if err == nil {
				t.Fatal("An HTTPRoute without the VAP parameter was accepted")
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
