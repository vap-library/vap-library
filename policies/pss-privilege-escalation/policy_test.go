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
var containerDeploymentYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox-deployment-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: privilege-escalation-%s
        image: busybox:1.28
        securityContext:
          allowPrivilegeEscalation: %s
`

var initContainerDeploymentYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: init-busybox-deployment-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
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

var containerRSYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: busybox-replicaset-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: privilege-escalation-%s
        image: busybox:1.28
        securityContext:
          allowPrivilegeEscalation: %s
`

var initContainerRSYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: init-busybox-replicaset-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
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
var containerDSYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: busybox-daemonset-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: privilege-escalation-%s
        image: busybox:1.28
        securityContext:
          allowPrivilegeEscalation: %s
`

var initContainerDSYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: init-busybox-daemonset-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
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

var containerSSYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: busybox-statefulset-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: privilege-escalation-%s
        image: busybox:1.28
        securityContext:
          allowPrivilegeEscalation: %s
`
var initContainerSSYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: init-busybox-statefulset-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
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
		// POD TESTS
		Assess("Successful deployment of a Pod with container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with init container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// DEPLOYMENT TESTS
		Assess("Successful deployment of a Deployment with container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// REPLICASET TESTS
		Assess("Successful deployment of a ReplicaSet with container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}
			
			return ctx
		}).
		// DAEMONSET TESTS
		Assess("Successful deployment of a DaemonSet with container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// STATEFULSET TESTS
		Assess("Successful deployment of a StatefulSet with container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with initContainer as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with initContainer as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})
		// ToDo: Add tests for ephemeral containers if possible

		// ToDo: Add tests for further non-pod objects

	_ = testEnv.Test(t, f.Feature())

}
