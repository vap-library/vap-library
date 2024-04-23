package pss_privilege_escalation

import (
	"context"
	"fmt"
	"log"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
	"time"
	"vap-library/testutils"

	"sigs.k8s.io/e2e-framework/pkg/env"
)

// test YAML locally
// simple test to attempt to deploy a container, run 2 - 1 aPE == false, 2 aPE == true
var containerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: privilege-escalation
spec:
  containers:
  - name: privilege-escalation
    image: busybox:1.28
    securityContext:
      allowPrivilegeEscalation: %s
`

var initContainerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-privilege-escalation
spec:
  initContainers:
  - name: init-privilege-escalation
    image: busybox:1.28
	securityContext:
	  allowPrivilegeEscalation: %s
`

var ephemeralContainerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-privilege-escalation
spec:
  ephemeralContainers:
  - name: init-privilege-escalation
    image: busybox:1.28
	securityContext:
	  allowPrivilegeEscalation: %s
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-privilege-escalation": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestPrivilegeEscalation(t *testing.T) {

	f := features.New("Privilege Escalation tests").
		Assess("Successful deployment of a pod as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			// namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, false))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a pod as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			// namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, true))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of an init container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			// namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, false))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of an init container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			// namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, true))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of an ephemeral container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			// namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ephemeralContainerYAML, false))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of an ephemeral container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			// namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ephemeralContainerYAML, true))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
