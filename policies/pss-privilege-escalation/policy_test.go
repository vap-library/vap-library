package pss_privilege_escalation

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
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"sigs.k8s.io/e2e-framework/pkg/env"
)

// TEST DATA FOR POD TESTS

var containerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: privilege-escalation-%s
  namespace: %s
spec:
  containers:
  - name: privilege-escalation-%s
    image: public.ecr.aws/docker/library/busybox:1.36
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
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      allowPrivilegeEscalation: %s
  initContainers:
  - name: init-privilege-escalation-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      allowPrivilegeEscalation: %s
`

// TEST DATA FOR DEPLOYMENT TESTS

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
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
      initContainers:
      - name: init-privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
`

// TEST DATA FOR REPLICASET TESTS

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
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
      initContainers:
      - name: init-privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
`

// TEST DATA FOR DAEMONSET TESTS

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
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
      initContainers:
      - name: init-privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
`

// TEST DATA FOR STATEFULSET TESTS

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
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
      initContainers:
      - name: init-privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
`

// TEST DATA FOR JOB TESTS

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
      - name: privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
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
      - name: privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
      initContainers:
      - name: init-privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
      restartPolicy: Never
`

// TEST DATA FOR CRONJOB TESTS

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
          - name: privilege-escalation-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              allowPrivilegeEscalation: %s
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
          - name: privilege-escalation-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              allowPrivilegeEscalation: %s
          initContainers:
          - name: init-privilege-escalation-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              allowPrivilegeEscalation: %s
          restartPolicy: OnFailure
`

// TEST DATA FOR REPLICATIONCONTROLLER TESTS

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
      - name: privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
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
      - name: privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
      initContainers:
      - name: init-privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
`

// TEST DATA FOR PODTEMPLATE TESTS

var containerPodTemplateYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: busybox-podtemplate-%s
  namespace: %s
template:
  spec:
    containers:
      - name: privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
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
      - name: privilege-escalation-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          allowPrivilegeEscalation: %s
    restartPolicy: Always
    initContainers:
    - name: init-privilege-escalation-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        allowPrivilegeEscalation: %s
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
           "allowPrivilegeEscalation": %s
         },
         "stdin": true,
         "targetContainerName": "privilege-escalation-ephemeral",
         "terminationMessagePolicy": "File",
         "tty": true
      }
    ]
  }
}
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-privilege-escalation": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil)
	if err != nil {
		log.Fatalf("Unable to create Kind cluster for test. Error msg: %s", err)
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
		}).
		// JOB TESTS
		Assess("Successful deployment of a Job with container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// CRONJOB TESTS
		Assess("Successful deployment of a CronJob with container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// REPLICATIONCONTROLLER TESTS
		Assess("Successful deployment of a ReplicationController with container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// PODTEMPLATE TESTS
		Assess("Successful deployment of a PodTemplate with container as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "success", namespace, "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "rejected", namespace, "rejected", "true"))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as allowPrivilegeEscalation is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "success", namespace, "success", "false", "success", "false"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as allowPrivilegeEscalation is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "rejected", namespace, "rejected", "false", "rejected", "true"))
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
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "ephemeral", namespace, "ephemeral", "false"))
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
			err = client.Resources(namespace).Get(ctx, "privilege-escalation-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "true"))

			patch := k8s.Patch{patchType, patchData}

			// patch the pod, this should FAIL!
			err = client.Resources(namespace).PatchSubresource(ctx, pod, "ephemeralcontainers", patch)
			if err == nil {
				t.Fatal("ephemeral container without securityContext.allowPrivilegeEscalation field should be rejected")
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
			err = client.Resources(namespace).Get(ctx, "privilege-escalation-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "false"))

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
