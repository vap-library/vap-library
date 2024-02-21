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
	// kindClusterName := envconf.RandomName("vap-library-", 16)
	// testenv.Setup(
	// 	envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
	// )

	// testenv.Finish(
	// 	envfuncs.DestroyCluster(kindClusterName),
	// )

	os.Exit(testenv.Run(m))
}

func TestHttpRouteEnforceHostnames(t *testing.T) {
	testutils.CreateFromFile("https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.5.1/experimental-install.yaml", t) // enable kubernetes gateway api
	testutils.CreateFromFile("policy.yaml", t)
	testutils.CreateFromFile("crd-parameter.yaml", t)

	t.Run("HTTPRoute with allowed routes should be allowed", func(t *testing.T) {
		testutils.DeleteNamespace("httproute-enforce-hostnames-vap-library-test0", t)
		testutils.CreationShouldSucceed(t, testutils.Dedent(`
			apiVersion: v1
			kind: Namespace
			metadata:
				labels:
					vap-library.com/httproute-enforce-hostnames: deny
				name: httproute-enforce-hostnames-vap-library-test01`))

		testutils.CreationShouldSucceed(t, testutils.Dedent(`
			apiVersion: vap-library.com/v1beta1
			kind: HTTPRouteEnforceHostnamesParam
			metadata:
				name: httproute-enforce-hostnames-vap-library-test
				namespace: httproute-enforce-hostnames-vap-library-test01
			spec:
				allowedHostnames:
				- test.example.com
				- test2.example.com`))
		testutils.CreateFromFile("binding.yaml", t)

		testutils.CreationShouldSucceed(t, testutils.Dedent(`
			apiVersion: gateway.networking.k8s.io/v1beta1
			kind: HTTPRoute
			metadata:
				name: httproute-example-vap-library-test-test01
				namespace: httproute-enforce-hostnames-vap-library-test01
			spec:
				hostnames:
				- test.example.com
				parentRefs:
				- name: dummy`))
	})

	t.Run("HTTPRoute with forbidden routes should be denied", func(t *testing.T) {
		testutils.DeleteNamespace("httproute-enforce-hostnames-vap-library-test02", t)
		testutils.CreationShouldSucceed(t, testutils.Dedent(`
			apiVersion: v1
			kind: Namespace
			metadata:
				labels:
					vap-library.com/httproute-enforce-hostnames: deny
				name: httproute-enforce-hostnames-vap-library-test02`))

		testutils.CreationShouldSucceed(t, testutils.Dedent(`
			apiVersion: vap-library.com/v1beta1
			kind: HTTPRouteEnforceHostnamesParam
			metadata:
				name: httproute-enforce-hostnames-vap-library-test
				namespace: httproute-enforce-hostnames-vap-library-test02
			spec:
				allowedHostnames:
				- test.example.com
				- test2.example.com`))
		testutils.CreateFromFile("binding.yaml", t)

		errorMessage := testutils.CreationShouldFail(t, testutils.Dedent(`
			apiVersion: gateway.networking.k8s.io/v1beta1
			kind: HTTPRoute
			metadata:
				name: httproute-example-vap-library-test-test02
				namespace: httproute-enforce-hostnames-vap-library-test02
			spec:
				hostnames:
				- notallowed.example.com
				parentRefs:
				- name: dummy`))

		if !strings.HasSuffix(errorMessage, "spec.hostnames must be present and each item must be on the spec.allowedHostnames list in the policy parameter\n") {
			t.Errorf("Unexpected error message: %s", errorMessage)
		}
	})
}
