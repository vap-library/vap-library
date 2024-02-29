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
apiVersion: gateway.networking.k8s.io/v1beta1
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
apiVersion: gateway.networking.k8s.io/v1beta1
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

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/httproute-enforce-hostnames": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", true, namespaceLabels)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	os.Exit(testEnv.Run(m))
}

func TestVAPHTTPRouteEnforceHostnamesValidHostname(t *testing.T) {

	f := features.New("HTTPRoute is accepted").
		Assess("A valid HTTPRoute is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// Get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)
			t.Logf("namespace: %s", namespace)
			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
