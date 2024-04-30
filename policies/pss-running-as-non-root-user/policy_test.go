// Defined tests for a single container with secctx and also multiple containers.

package pss_running_as_non_root_user

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
  name: running-as-non-root-user-%s
  namespace: %s
spec:
  containers:
  - name: running-as-non-root-user-%s
    image: busybox:1.28
    securityContext:
      runAsUser: %s
`

var multipleContainerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: running-as-non-root-user-%s
  namespace: %s
spec:
  containers:
  - name: running-as-non-root-user-%s
    image: busybox:1.28
    securityContext:
      runAsUser: %s
  - name: running-as-non-root-user-%s
    image: busybox:1.28
    securityContext:
      runAsUser: %s
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-running-as-non-root-user": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}


func TestRunningAsNonRoot(t *testing.T) {

	f := features.New("Running as Non-Root tests").
		// POD TESTS
		Assess("Successful deployment of a Pod with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "success", namespace, "success", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with 2 containers as runAsUser not set to 0 on either container", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(multipleContainerYAML, "success-multi", namespace, "success", "100", "success-2", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with 2 containers as runAsUser is set to 0 on one container", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(multipleContainerYAML, "rejected-multi", namespace, "rejected", "0", "rejected-2", "10"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})

		_ = testEnv.Test(t, f.Feature())

	}