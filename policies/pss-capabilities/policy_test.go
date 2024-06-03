package pss_capabilities

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
  name: capabilities-%s
  namespace: %s
spec:
  containers:
  - name: capabilities
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      capabilities:
        drop:
        - %s
        add:
        - %s
`

var containerNoAddYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: capabilities-no-add-%s
  namespace: %s
spec:
  containers:
  - name: capabilities
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      capabilities:
        drop:
        - %s
`

var initContainerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-capabilities-%s
  namespace: %s
spec:
  containers:
  - name: capabilities
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      capabilities:
        drop:
        - %s
        add:
        - %s
  initContainers:
  - name: init-capabilities
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      capabilities:
        drop:
        - %s
        add:
        - %s
`

var initContainerNoAddYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-capabilities-no-add-%s
  namespace: %s
spec:
  containers:
  - name: capabilities
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      capabilities:
        drop:
        - %s
        add:
        - %s
  initContainers:
  - name: init-capabilities
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      capabilities:
        drop:
        - %s
`

// TEST DATA FOR DEPLOYMENT TESTS

var containerDeploymentYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: capabilities-deployment-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var containerDeploymentNoAddYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: capabilities-deployment-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

var initContainerDeploymentYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: init-capabilities-deployment-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var initContainerDeploymentNoAddYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: init-capabilities-deployment-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

// TEST DATA FOR REPLICASET TESTS

var containerRSYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: capabilities-replicaset-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var containerRSNoAddYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: capabilities-replicaset-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

var initContainerRSYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: init-capabilities-replicaset-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var initContainerRSNoAddYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: init-capabilities-replicaset-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

// TEST DATA FOR DAEMONSET TESTS

var containerDSYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: capabilities-daemonset-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var containerDSNoAddYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: capabilities-daemonset-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

var initContainerDSYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: init-capabilities-daemonset-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var initContainerDSNoAddYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: init-capabilities-daemonset-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

// TEST DATA FOR STATEFULSET TESTS

var containerSSYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: capabilities-statefulset-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var containerSSNoAddYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: capabilities-statefulset-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

var initContainerSSYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: init-capabilities-statefulset-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var initContainerSSNoAddYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: init-capabilities-statefulset-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

// TEST DATA FOR JOB TESTS

var containerJobYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: capabilities-job-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      restartPolicy: Never
`

var containerJobNoAddYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: capabilities-job-no-add-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
      restartPolicy: Never
`

var initContainerJobYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: init-capabilities-job-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      restartPolicy: Never
`

var initContainerJobNoAddYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: init-capabilities-job-no-add-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
      restartPolicy: Never
`

// TEST DATA FOR CRONJOB TESTS

var containerCronJobYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: capabilities-cronjob-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: capabilities
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              capabilities:
                drop:
                - %s
                add:
                - %s
          restartPolicy: OnFailure
`

var containerCronJobNoAddYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: capabilities-cronjob-no-add-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: capabilities
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              capabilities:
                drop:
                - %s
          restartPolicy: OnFailure
`

var initContainerCronJobYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: init-capabilities-cronjob-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: capabilities
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              capabilities:
                drop:
                - %s
                add:
                - %s
          initContainers:
          - name: init-capabilities
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              capabilities:
                drop:
                - %s
                add:
                - %s
          restartPolicy: OnFailure
`

var initContainerCronJobNoAddYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: init-capabilities-cronjob-no-add-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: capabilities
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              capabilities:
                drop:
                - %s
                add:
                - %s
          initContainers:
          - name: init-capabilities
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              capabilities:
                drop:
                - %s
          restartPolicy: OnFailure
`

// TEST DATA FOR REPLICATIONCONTROLLER TESTS

var containerRCYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: capabilities-replicationcontroller-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var containerRCNoAddYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: capabilities-replicationcontroller-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

var initContainerRCYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: init-capabilities-replicationcontroller-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var initContainerRCNoAddYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: init-capabilities-replicationcontroller-no-add-%s
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
      - name: capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
      initContainers:
      - name: init-capabilities
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          capabilities:
            drop:
            - %s
`

// TEST DATA FOR PODTEMPLATE TESTS

var containerPodTemplateYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: capabilities-podtemplate-%s
  namespace: %s
template:
  spec:
    containers:
    - name: capabilities
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
    restartPolicy: Always
`

var containerPodTemplateNoAddYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: capabilities-podtemplate-no-add-%s
  namespace: %s
template:
  spec:
    containers:
    - name: capabilities
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
          capabilities:
            drop:
            - %s
    restartPolicy: Always
`

var initContainerPodTemplateYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: init-capabilities-podtemplate-%s
  namespace: %s
template:
  spec:
    containers:
    - name: capabilities
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
    restartPolicy: Always
    initContainers:
    - name: init-capabilities
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
          capabilities:
            drop:
            - %s
            add:
            - %s
`

var initContainerPodTemplateNoAddYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: init-capabilities-podtemplate-no-add-%s
  namespace: %s
template:
  spec:
    containers:
    - name: capabilities
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        capabilities:
          drop:
          - %s
          add:
          - %s
    restartPolicy: Always
    initContainers:
    - name: init-capabilities
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        capabilities:
          drop:
          - %s
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
           "capabilities": {
            "drop": [
              "%s"
            ],
            "add": [
              "%s"
            ]
           }
         },
         "stdin": true,
         "targetContainerName": "capabilities",
         "terminationMessagePolicy": "File",
         "tty": true
      }
    ]
  }
}
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-capabilities": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil)
	if err != nil {
		log.Fatalf("Unable to create Kind cluster for test. Error msg: %s", err)
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestCapabilities(t *testing.T) {

	f := features.New("Capabilities tests").
		// POD TESTS
		Assess("Successful deployment of a Pod with container as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "success", namespace, "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with container as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerNoAddYAML, "success", namespace, "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with initContainer as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with initContainer as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerNoAddYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "rejected", namespace, "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerNoAddYAML, "rejected", namespace, "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerNoAddYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		// DEPLOYMENT TESTS
		Assess("Successful deployment of a Deployment with container as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "success", namespace, "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with container as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentNoAddYAML, "success", namespace, "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentNoAddYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "rejected", namespace, "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentNoAddYAML, "rejected", namespace, "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentNoAddYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		// REPLICASET TESTS
		Assess("Successful deployment of a ReplicaSet with container as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "success", namespace, "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with container as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSNoAddYAML, "success", namespace, "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSNoAddYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "rejected", namespace, "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSNoAddYAML, "rejected", namespace, "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSNoAddYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		// DAEMONSET TESTS
		Assess("Successful deployment of a DaemonSet with container as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "success", namespace, "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with container as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSNoAddYAML, "success", namespace, "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSNoAddYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "rejected", namespace, "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSNoAddYAML, "rejected", namespace, "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSNoAddYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		// STATEFULSET TESTS
		Assess("Successful deployment of a StatefulSet with container as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "success", namespace, "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with container as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSNoAddYAML, "success", namespace, "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with initContainer as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with initContainer as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSNoAddYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with container as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "rejected", namespace, "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with container as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSNoAddYAML, "rejected", namespace, "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with initContainer as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with initContainer as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSNoAddYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		// JOB TESTS
		Assess("Successful deployment of a Job with container as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "success", namespace, "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with container as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobNoAddYAML, "success", namespace, "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobNoAddYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "rejected", namespace, "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a Job with container as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobNoAddYAML, "rejected", namespace, "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobNoAddYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		// CRONJOB TESTS
		Assess("Successful deployment of a CronJob with container as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "success", namespace, "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with container as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobNoAddYAML, "success", namespace, "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobNoAddYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "rejected", namespace, "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobNoAddYAML, "rejected", namespace, "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobNoAddYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		// REPLICATIONCONTROLLER TESTS
		Assess("Successful deployment of a ReplicationController with container as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "success", namespace, "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with container as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCNoAddYAML, "success", namespace, "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCNoAddYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "rejected", namespace, "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCNoAddYAML, "rejected", namespace, "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCNoAddYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		// PODTEMPLATE TESTS
		Assess("Successful deployment of a PodTemplate with container as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "success", namespace, "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with container as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateNoAddYAML, "success", namespace, "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as ALL in capabilities.deny and capabilities.add only includes NET_BIND_SERVICE", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NET_BIND_SERVICE"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as ALL in capabilities.deny and no disallowed values in capabilites.add", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateNoAddYAML, "success", namespace, "ALL", "NET_BIND_SERVICE", "ALL"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "rejected", namespace, "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateNoAddYAML, "rejected", namespace, "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as disallowed capability added", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "ALL", "NOT_ALLOWED"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
			}
			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as ALL capabilities not dropped", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateNoAddYAML, "rejected", namespace, "ALL", "NET_BIND_SERVICE", "NONE"))
			if err == nil {
				t.Fatal("containers with invalid capabilities were accepted.")
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
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "ephemeral", namespace, "ALL", "NET_BIND_SERVICE"))
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
			err = client.Resources(namespace).Get(ctx, "capabilities-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "CAP_NET_RAW", "NET_BIND_SERVICE"))

			patch := k8s.Patch{patchType, patchData}

			// patch the pod, this should FAIL!
			err = client.Resources(namespace).PatchSubresource(ctx, pod, "ephemeralcontainers", patch)
			if err == nil {
				t.Fatal("ephemeral container without securityContext.Capabilities.Drop['ALL'] should be rejected")
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
			err = client.Resources(namespace).Get(ctx, "capabilities-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "ALL", "NET_BIND_SERVICE"))

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
