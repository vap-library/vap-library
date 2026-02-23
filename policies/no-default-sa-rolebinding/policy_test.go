package no_default_sa_rolebinding

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"vap-library/testutils"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var roleBindingValidYAML string = `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: valid-rolebinding
  namespace: %s
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: example-role
subjects:
- kind: ServiceAccount
  name: my-service-account
  namespace: %s
`

var roleBindingDefaultSAYAML string = `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: invalid-rolebinding
  namespace: %s
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: example-role
subjects:
- kind: ServiceAccount
  name: default
  namespace: %s
`

var roleBindingMixedSubjectsYAML string = `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: mixed-rolebinding
  namespace: %s
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: example-role
subjects:
- kind: ServiceAccount
  name: my-service-account
  namespace: %s
- kind: ServiceAccount
  name: default
  namespace: %s
`

var roleBindingNoSubjectsYAML string = `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: no-subjects-rolebinding
  namespace: %s
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: example-role
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/no-default-sa-rolebinding": "deny"}
	var bindingsToGenerate = map[string]bool{"no-default-sa-rolebinding": false}
	var err error

	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil, bindingsToGenerate)
	if err != nil {
		log.Fatalf("Unable to create Kind cluster for test. Error msg: %s", err)
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestRoleBindingNoDefaultSA(t *testing.T) {

	f := features.New("RoleBinding default service account tests").
		Assess("A RoleBinding with a non-default service account is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(roleBindingValidYAML, namespace, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("A RoleBinding with the default service account is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(roleBindingDefaultSAYAML, namespace, namespace))
			if err == nil {
				t.Fatal("A RoleBinding with the default service account was accepted")
			}

			return ctx
		}).
		Assess("A RoleBinding with mixed subjects including the default service account is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(roleBindingMixedSubjectsYAML, namespace, namespace, namespace))
			if err == nil {
				t.Fatal("A RoleBinding with mixed subjects including the default service account was accepted")
			}

			return ctx
		}).
		Assess("A RoleBinding without subjects is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(roleBindingNoSubjectsYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
