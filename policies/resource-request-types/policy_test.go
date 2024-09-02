package resource_request_types

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"vap-library/testutils"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"sigs.k8s.io/e2e-framework/pkg/env"
)

// TEST DATA FOR POD TESTS

var testParameterYAML string = `
apiVersion: vap-library.com/v1beta1
kind: VAPLibResourceRequestTypesParam
metadata:
  name: resource-request-types.vap-library.com
  namespace: %s
spec:
  enforcedResourceRequestTypes:
  - cpu
  - memory
  - ephemeral-storage
`

var containerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: resource-request-types-%s
  namespace: %s
spec:
  containers:
  - name: resource-request-types-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    resources:
      requests:
        %s
`

var initContainerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-resource-request-types-%s
  namespace: %s
spec:
  containers:
  - name: resource-request-types-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    resources:
      requests:
        %s
  initContainers:
  - name: init-resource-request-types-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    resources:
      requests:
        %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
      initContainers:
      - name: init-resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
      initContainers:
      - name: init-resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
      initContainers:
      - name: init-resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
      initContainers:
      - name: init-resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
      initContainers:
      - name: init-resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
          - name: resource-request-types-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            resources:
              requests:
                %s
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
          - name: resource-request-types-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            resources:
              requests:
                %s
          initContainers:
          - name: init-resource-request-types-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            resources:
              requests:
                %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
      - name: resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
      initContainers:
      - name: init-resource-request-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        resources:
          requests:
            %s
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
    - name: resource-request-types-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      resources:
        requests:
          %s
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
    - name: resource-request-types-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      resources:
        requests:
          %s
    restartPolicy: Always
    initContainers:
    - name: init-resource-request-types-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      resources:
        requests:
          %s
`

var matchingResourceRequests string = `cpu: 500m
        memory: 128Mi
        ephemeral-storage: 400M`

var nonMatchingResourceRequests string = `cpu: 500m
        memory: 128Mi`

var matchingResourceRequestsWorkload string = `cpu: 500m
            memory: 128Mi
            ephemeral-storage: 400M`

var nonMatchingResourceRequestsWorkload string = `cpu: 500m
            memory: 128Mi`

var matchingResourceRequestsCronjob string = `cpu: 500m
                memory: 128Mi
                ephemeral-storage: 400M`

var nonMatchingResourceRequestsCronjob string = `cpu: 500m
                memory: 128Mi`

var matchingResourceRequestsPodTemplate string = `cpu: 500m
          memory: 128Mi
          ephemeral-storage: 400M`

var nonMatchingResourceRequestsPodTemplate string = `cpu: 500m
          memory: 128Mi`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/resource-request-types": "deny"}
	var bindingsToGenerate = map[string]bool{"resource-request-types": true}
	var err error

	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil, bindingsToGenerate)
	if err != nil {
		log.Fatalf("Unable to create Kind cluster for test. Error msg: %s", err)
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestResourceRequests(t *testing.T) {

	f := features.New("Resource request tests with parameter").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// apply parameter first
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(testParameterYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the parameter to be registered properly
			time.Sleep(10 * time.Second)

			return ctx
		}).
		// POD TESTS
		Assess("Successful deployment of a Pod with container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "success", namespace, "success", matchingResourceRequests))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "rejected", namespace, "rejected", nonMatchingResourceRequests))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with init container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "success", namespace, "success", matchingResourceRequests, "success", matchingResourceRequests))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "rejected", namespace, "rejected", matchingResourceRequests, "rejected", nonMatchingResourceRequests))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// DEPLOYMENT TESTS
		Assess("Successful deployment of a Deployment with container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "success", namespace, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "rejected", namespace, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "success", namespace, "success", matchingResourceRequestsWorkload, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "rejected", namespace, "rejected", matchingResourceRequestsWorkload, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// REPLICASET TESTS
		Assess("Successful deployment of a ReplicaSet with container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "success", namespace, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "rejected", namespace, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "success", namespace, "success", matchingResourceRequestsWorkload, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "rejected", namespace, "rejected", matchingResourceRequestsWorkload, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// DAEMONSET TESTS
		Assess("Successful deployment of a DaemonSet with container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "success", namespace, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "rejected", namespace, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "success", namespace, "success", matchingResourceRequestsWorkload, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "rejected", namespace, "rejected", matchingResourceRequestsWorkload, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// STATEFULSET TESTS
		Assess("Successful deployment of a StatefulSet with container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "success", namespace, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with container as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "rejected", namespace, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with initContainer as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "success", namespace, "success", matchingResourceRequestsWorkload, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with initContainer as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "rejected", namespace, "rejected", matchingResourceRequestsWorkload, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// JOB TESTS
		Assess("Successful deployment of a Job with container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "success", namespace, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "rejected", namespace, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "success", namespace, "success", matchingResourceRequestsWorkload, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "rejected", namespace, "rejected", matchingResourceRequestsWorkload, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// CRONJOB TESTS
		Assess("Successful deployment of a CronJob with container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "success", namespace, "success", matchingResourceRequestsCronjob))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "rejected", namespace, "rejected", nonMatchingResourceRequestsCronjob))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "success", namespace, "success", matchingResourceRequestsCronjob, "success", matchingResourceRequestsCronjob))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "rejected", namespace, "rejected", matchingResourceRequestsCronjob, "rejected", nonMatchingResourceRequestsCronjob))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// REPLICATIONCONTROLLER TESTS
		Assess("Successful deployment of a ReplicationController with container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "success", namespace, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "rejected", namespace, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "success", namespace, "success", matchingResourceRequestsWorkload, "success", matchingResourceRequestsWorkload))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "rejected", namespace, "rejected", matchingResourceRequestsWorkload, "rejected", nonMatchingResourceRequestsWorkload))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		// PODTEMPLATE TESTS
		Assess("Successful deployment of a PodTemplate with container as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "success", namespace, "success", matchingResourceRequestsPodTemplate))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "rejected", namespace, "rejected", nonMatchingResourceRequestsPodTemplate))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as parameter-specified request types are set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "success", namespace, "success", matchingResourceRequestsPodTemplate, "success", matchingResourceRequestsPodTemplate))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as parameter-specified request types are not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "rejected", namespace, "rejected", matchingResourceRequestsPodTemplate, "rejected", nonMatchingResourceRequestsPodTemplate))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
