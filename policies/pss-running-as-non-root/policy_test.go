// Policies have been defined, outstanding are the tests.
// Current tests defined:
// Only covered a pod with a containers block (no init or ephemeral) so need to expand for these use cases as well as all additional objects

package pss_running_as_non_root

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
  name: running-as-non-root-%s
  namespace: %s
spec:
  containers:
  - name: running-as-non-root-%s
    image: busybox:1.28
    securityContext:
      runAsNonRoot: %s
`

var containerWithDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: running-as-non-root-with-default-%s
  namespace: %s
spec:
  securityContext:
    runAsNonRoot: %s
  containers:
  - name: running-as-non-root-with-default-%s
    image: busybox:1.28
    securityContext:
      runAsNonRoot: %s
`

var containerOnlyDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: running-as-non-root-%s
  namespace: %s
spec:
  securityContext:
    runAsNonRoot: %s
  containers:
  - name: running-as-non-root-%s
    image: busybox:1.28
`

var initContainerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-running-as-non-root-%s
  namespace: %s
spec:
  containers:
  - name: running-as-non-root-%s
    image: busybox:1.28
    securityContext:
      runAsNonRoot: %s
  initContainers:
  - name: init-running-as-non-root-%s
    image: busybox:1.28
    securityContext:
      runAsNonRoot: %s
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
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
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
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
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
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
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
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
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
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
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
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
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
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
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
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
`
var containerJobYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: busybox-job-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
      restartPolicy: Never
`

var initContainerJobYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: init-busybox-job-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
      restartPolicy: Never
`

var containerCronJobYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: busybox-cronjob-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: running-as-non-root-%s
            image: busybox:1.28
            securityContext:
              runAsNonRoot: %s
          restartPolicy: OnFailure
`
var initContainerCronJobYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: init-busybox-cronjob-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: running-as-non-root-%s
            image: busybox:1.28
            securityContext:
              runAsNonRoot: %s
          initContainers:
          - name: init-running-as-non-root-%s
            image: busybox:1.28
            securityContext:
              runAsNonRoot: %s
          restartPolicy: OnFailure
`

var containerRCYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: busybox-replicationcontroller-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
`

var initContainerRCYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: init-busybox-replicationcontroller-%s
  namespace: %s
  labels:
    app: busybox
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
`
var containerPodTemplateYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: busybox-podtemplate-%s
  namespace: %s
template:
  spec:
    containers:
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
    restartPolicy: Always
`

var initContainerPodTemplateYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: init-busybox-podtemplate-%s
  namespace: %s
template:
  spec:
    containers:
      - name: running-as-non-root-%s
        image: busybox:1.28
        securityContext:
          runAsNonRoot: %s
    restartPolicy: Always
    initContainers:
    - name: init-running-as-non-root-%s
      image: busybox:1.28
      securityContext:
        runAsNonRoot: %s
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-running-as-non-root": "deny"}

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

	f := features.New("Running as Non-Root tests").
		// POD TESTS
		Assess("Successful deployment of a Pod with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "success", namespace, "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "rejected", namespace, "rejected", "false"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "success-default-true", namespace, "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "success-default-false", namespace, "false", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with container as container.runAsNonRoot is not defined and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerOnlyDefaultYAML, "success-only-default-true", namespace, "true", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})
		// Assess("Successful deployment of a Pod with init container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "success", namespace, "success", "false", "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a Pod with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // DEPLOYMENT TESTS
		// Assess("Successful deployment of a Deployment with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "success", namespace, "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a Deployment with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "rejected", namespace, "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Successful deployment of a Deployment with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "success", namespace, "success", "false", "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a Deployment with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // REPLICASET TESTS
		// Assess("Successful deployment of a ReplicaSet with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "success", namespace, "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicaSet with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "rejected", namespace, "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Successful deployment of a ReplicaSet with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "success", namespace, "success", "false", "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicaSet with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}
			
		// 	return ctx
		// }).
		// // DAEMONSET TESTS
		// Assess("Successful deployment of a DaemonSet with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "success", namespace, "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a DaemonSet with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "rejected", namespace, "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Successful deployment of a DaemonSet with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "success", namespace, "success", "false", "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a DaemonSet with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // STATEFULSET TESTS
		// Assess("Successful deployment of a StatefulSet with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "success", namespace, "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a StatefulSet with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "rejected", namespace, "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Successful deployment of a StatefulSet with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "success", namespace, "success", "false", "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a StatefulSet with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // JOB TESTS
		// Assess("Successful deployment of a Job with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "success", namespace, "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a Job with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "rejected", namespace, "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Successful deployment of a Job with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "success", namespace, "success", "false", "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a Job with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // CRONJOB TESTS
		// Assess("Successful deployment of a CronJob with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "success", namespace, "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a CronJob with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "rejected", namespace, "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Successful deployment of a CronJob with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "success", namespace, "success", "false", "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a CronJob with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // REPLICATIONCONTROLLER TESTS
		// Assess("Successful deployment of a ReplicationController with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "success", namespace, "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicationController with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "rejected", namespace, "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Successful deployment of a ReplicationController with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "success", namespace, "success", "false", "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicationController with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // PODTEMPLATE TESTS
		// Assess("Successful deployment of a PodTemplate with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "success", namespace, "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a PodTemplate with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "rejected", namespace, "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Successful deployment of a PodTemplate with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "success", namespace, "success", "false", "success", "false"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a PodTemplate with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// })
		// ToDo: Add tests for ephemeral containers if possible

		// ToDo: Add tests for further non-pod objects

	_ = testEnv.Test(t, f.Feature())

}