// Defined tests for a single container with secctx and also multiple containers.

package pss_running_as_non_root_user

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"vap-library/testutils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// TEST DATA FOR POD TESTS

var containerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: running-as-non-root-user-%s
  namespace: %s
spec:
  containers:
  - name: running-as-non-root-user-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsUser: %s
`

var containerNoRunAsUserYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: running-as-non-root-user-undefined-%s
  namespace: %s
spec:
  containers:
  - name: running-as-non-root-user-%s
    image: public.ecr.aws/docker/library/busybox:1.36
`

var containerWithDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: running-as-non-root-user-with-default-%s
  namespace: %s
spec:
  securityContext:
    runAsUser: %s
  containers:
  - name: running-as-non-root-user-with-default-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsUser: %s
`

var initContainerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-running-as-non-root-user-%s
  namespace: %s
spec:
  containers:
  - name: running-as-non-root-user-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsUser: %s
  initContainers:
  - name: init-running-as-non-root-user-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsUser: %s
`

// TEST DATA FOR DEPLOYMENT TESTS

var containerDeploymentYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var containerDeploymentNoRunAsUserYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-running-as-non-root-user-undefined-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

var containerDeploymentWithDefaultYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-running-as-non-root-user-with-default-%s
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
      securityContext:
        runAsUser: %s
      containers:
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var initContainerDeploymentYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: init-deployment-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
      initContainers:
      - name: init-running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

// TEST DATA FOR REPLICASET TESTS

var containerRSYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var containerRSNoRunAsUserYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-running-as-non-root-user-undefined-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

var containerRSWithDefaultYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-running-as-non-root-user-with-default-%s
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
      securityContext:
        runAsUser: %s
      containers:
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var initContainerRSYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: init-replicaset-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
      initContainers:
      - name: init-running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

// TEST DATA FOR DAEMONSET TESTS

var containerDSYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var containerDSNoRunAsUserYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-running-as-non-root-user-undefined-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

var containerDSWithDefaultYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-running-as-non-root-user-with-default-%s
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
      securityContext:
        runAsUser: %s
      containers:
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var initContainerDSYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: init-daemonset-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
      initContainers:
      - name: init-running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

// TEST DATA FOR STATEFULSET TESTS

var containerSSYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var containerSSNoRunAsUserYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-running-as-non-root-user-undefined-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

var containerSSWithDefaultYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-running-as-non-root-user-with-default-%s
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
      securityContext:
        runAsUser: %s
      containers:
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var initContainerSSYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: init-statefulset-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
      initContainers:
      - name: init-running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

// TEST DATA FOR JOB TESTS

var containerJobYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-running-as-non-root-user-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
      restartPolicy: Never
`

var containerJobNoRunAsUserYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-running-as-non-root-user-undefined-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      restartPolicy: Never
`

var containerJobWithDefaultYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-running-as-non-root-user-with-default-%s
  namespace: %s
spec:
  template:
    spec:
      securityContext:
        runAsUser: %s
      containers:
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
      restartPolicy: Never
`

var initContainerJobYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: init-job-running-as-non-root-user-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
      initContainers:
      - name: init-running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
      restartPolicy: Never
`

// TEST DATA FOR CRONJOB TESTS

var containerCronJobYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-running-as-non-root-user-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: running-as-non-root-user-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsUser: %s
          restartPolicy: OnFailure
`

var containerCronJobNoRunAsUserYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-running-as-non-root-user-undefined-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: running-as-non-root-user-%s
            image: public.ecr.aws/docker/library/busybox:1.36
          restartPolicy: OnFailure
`

var containerCronJobWithDefaultYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-running-as-non-root-user-with-default-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          securityContext:
            runAsUser: %s
          containers:
          - name: running-as-non-root-user-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsUser: %s
          restartPolicy: OnFailure
`

var initContainerCronJobYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: init-cronjob-running-as-non-root-user-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: running-as-non-root-user-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsUser: %s
          initContainers:
          - name: init-running-as-non-root-user-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsUser: %s
          restartPolicy: OnFailure
`

// TEST DATA FOR REPLICATIONCONTROLLER TESTS

var containerRCYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var containerRCNoRunAsUserYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-running-as-non-root-user-undefined-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

var containerRCWithDefaultYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-running-as-non-root-user-with-default-%s
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
      securityContext:
        runAsUser: %s
      containers:
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

var initContainerRCYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: init-replicationcontroller-running-as-non-root-user-%s
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
      - name: running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
      initContainers:
      - name: init-running-as-non-root-user-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsUser: %s
`

// TEST DATA FOR PODTEMPLATE TESTS

var containerPodTemplateYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-running-as-non-root-user-%s
  namespace: %s
template:
  spec:
    containers:
    - name: running-as-non-root-user-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsUser: %s
    restartPolicy: Always
`

var containerPodTemplateNoRunAsUserYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-running-as-non-root-user-undefined-%s
  namespace: %s
template:
  spec:
    containers:
    - name: running-as-non-root-user-%s
      image: public.ecr.aws/docker/library/busybox:1.36
    restartPolicy: Always
`

var containerPodTemplateWithDefaultYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-running-as-non-root-user-with-default%s
  namespace: %s
template:
  spec:
    securityContext:
      runAsUser: %s
    containers:
    - name: running-as-non-root-user-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsUser: %s
    restartPolicy: Always
`

var initContainerPodTemplateYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: init-podtemplate-running-as-non-root-user-%s
  namespace: %s
template:
  spec:
    containers:
    - name: running-as-non-root-user-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsUser: %s
    restartPolicy: Always
    initContainers:
    - name: init-running-as-non-root-user-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsUser: %s
`

// TEST DATA FOR PATCHING A POD TO ADD AN EPHEMERAL CONTAINER

var containerEphemeralPatchYAML string = `
{
  "spec": {
    "ephemeralContainers": [
      {
         "image": "public.ecr.aws/docker/library/busybox:1.36",
         "name": "ephemeral",
         "resources": {},
         "securityContext": {
           "runAsUser": %s
         },
         "stdin": true,
         "targetContainerName": "running-as-non-root-user-ephemeral",
         "terminationMessagePolicy": "File",
         "tty": true
      }
    ]
  }
}
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-running-as-non-root-user": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil)
	if err != nil {
		log.Fatalf("Unable to create Kind cluster for test. Error msg: %s", err)
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
		Assess("Successful deployment of a Pod with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerNoRunAsUserYAML, "success", namespace, "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// DEPLOYMENT TESTS
		Assess("Successful deployment of a Deployment with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "success", namespace, "success", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Assess("Successful deployment of a Deployment with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// get namespace
		namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// this should PASS!
		err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentNoRunAsUserYAML, "success", namespace, "success"))
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}).
		Assess("Rejected deployment of a Deployment with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// REPLICASET TESTS
		Assess("Successful deployment of a ReplicaSet with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "success", namespace, "success", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Assess("Successful deployment of a ReplicaSet with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// get namespace
		namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// this should PASS!
		err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSNoRunAsUserYAML, "success", namespace, "success"))
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}).
		Assess("Rejected deployment of a ReplicaSet with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// DAEMONSET TESTS
		Assess("Successful deployment of a DaemonSet with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "success", namespace, "success", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Assess("Successful deployment of a DaemonSet with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// get namespace
		namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// this should PASS!
		err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSNoRunAsUserYAML, "success", namespace, "success"))
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}).
		Assess("Rejected deployment of a DaemonSet with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// STATEFULSET TESTS
		Assess("Successful deployment of a StatefulSet with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "success", namespace, "success", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Assess("Successful deployment of a StatefulSet with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// get namespace
		namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// this should PASS!
		err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSNoRunAsUserYAML, "success", namespace, "success"))
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}).
		Assess("Rejected deployment of a StatefulSet with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// JOB TESTS
		Assess("Successful deployment of a Job with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "success", namespace, "success", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Assess("Successful deployment of a Job with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// get namespace
		namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// this should PASS!
		err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobNoRunAsUserYAML, "success", namespace, "success"))
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}).
		Assess("Rejected deployment of a Job with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// CRONJOB TESTS
		Assess("Successful deployment of a CronJob with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "success", namespace, "success", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Assess("Successful deployment of a CronJob with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// get namespace
		namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// this should PASS!
		err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobNoRunAsUserYAML, "success", namespace, "success"))
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}).
		Assess("Rejected deployment of a CronJob with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// REPLICATIONCONTROLLER TESTS
		Assess("Successful deployment of a ReplicationController with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "success", namespace, "success", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Assess("Successful deployment of a ReplicationController with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// get namespace
		namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// this should PASS!
		err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCNoRunAsUserYAML, "success", namespace, "success"))
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}).
		Assess("Rejected deployment of a ReplicationController with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// PODTEMPLATE TESTS
		Assess("Successful deployment of a PodTemplate with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "success", namespace, "success", "100"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Assess("Successful deployment of a PodTemplate with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// get namespace
		namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// this should PASS!
		err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateNoRunAsUserYAML, "success", namespace, "success"))
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}).
		Assess("Rejected deployment of a PodTemplate with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "rejected", namespace, "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestEphemeralContainers(t *testing.T) {

	f := features.New("Pods with ephemeral containers").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// create a pod that will be used for ephemeral container tests
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "ephemeral", namespace, "ephemeral", "100"))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the pod
			time.Sleep(2 * time.Second)

			return ctx
		}).
		Assess("An invalid ephemeral container is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// get client
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			// get the pod that was created in setup to attach an ephemeral container to it
			pod := &v1.Pod{}
			err = client.Resources(namespace).Get(ctx, "running-as-non-root-user-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "0"))

			patch := k8s.Patch{patchType, patchData}

			// patch the pod, this should FAIL!
			err = client.Resources(namespace).PatchSubresource(ctx, pod, "ephemeralcontainers", patch)
			if err == nil {
				t.Fatal("ephemeral container with securityContext.runAsUser field set to 0 should be rejected")
			}

			return ctx
		}).
		Assess("A valid ephemeral container is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// get client
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			// get the pod that was created in setup to attach an ephemeral container to it
			pod := &v1.Pod{}
			err = client.Resources(namespace).Get(ctx, "running-as-non-root-user-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "100"))

			patch := k8s.Patch{patchType, patchData}

			// patch the pod
			err = client.Resources(namespace).PatchSubresource(ctx, pod, "ephemeralcontainers", patch)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		})
	_ = testEnv.Test(t, f.Feature())

}
