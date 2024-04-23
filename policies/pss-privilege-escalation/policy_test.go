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

var containerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: privilege-escalation-%s
  namespace: %s
spec:
  containers:
  - name: privilege-escalation-%s
    image: busybox:1.28
    securityContext:
      allowPrivilegeEscalation: %s
`

var initContainerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-privilege-escalation-%s
  namespace: %s
spec:
  containers:
  - name: privilege-escalation-%s
    image: busybox:1.28
    securityContext:
      allowPrivilegeEscalation: %s
  initContainers:
  - name: init-privilege-escalation-%s
    image: busybox:1.28
    securityContext:
      allowPrivilegeEscalation: %s
`

// ToDo: Add test data for non-pod objects


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
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a pod as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of an init container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of an init container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})
		
		// ToDo: Add tests for ephemeral containers if possible

		// ToDo: Add tests for non-pod objects

	_ = testEnv.Test(t, f.Feature())

}
