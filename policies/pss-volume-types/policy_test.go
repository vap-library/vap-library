package pss_volume_types

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

var containerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: volume-types-pod-invalid
  namespace: %s
spec:
  containers:
  - name: volume-types
    image: public.ecr.aws/docker/library/busybox:1.36
  volumes:
  - name: example-volume
    hostPath:
      path: /data/foo # directory location on host
      type: Directory # this field is optional
`

var configMapYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: volume-types-pod-configmap
  namespace: %s
spec:
  containers:
  - name: volume-types
    image: public.ecr.aws/docker/library/busybox:1.36
  volumes:
  - name: example-volume
    configMap:
      name: log-config
      items:
      - key: log_level
        path: log_level
`

var downwardAPIYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: volume-types-pod-downwardapi
  namespace: %s
spec:
  containers:
  - name: volume-types
    image: public.ecr.aws/docker/library/busybox:1.36
  volumes:
  - name: example-volume
    downwardAPI:
    items:
    - path: "labels"
      fieldRef:
        fieldPath: metadata.labels
`

var emptyDirYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: volume-types-pod-emptydir
  namespace: %s
spec:
  containers:
  - name: volume-types
    image: public.ecr.aws/docker/library/busybox:1.36
  volumes:
  - name: example-volume
    emptyDir:
      sizeLimit: 500Mi
`

var emphemeralYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: volume-types-pod-ephemeral
  namespace: %s
spec:
  containers:
  - name: volume-types
    image: public.ecr.aws/docker/library/busybox:1.36
  volumes:
  - name: example-volume
    ephemeral:
      volumeClaimTemplate:
        metadata:
          labels:
            type: my-frontend-volume
        spec:
          accessModes: [ "ReadWriteOnce" ]
          storageClassName: "scratch-storage-class"
          resources:
            requests:
              storage: 1Gi
`

var pvcYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: volume-types-pod-pvc
  namespace: %s
spec:
  containers:
  - name: volume-types
    image: public.ecr.aws/docker/library/busybox:1.36
  volumes:
  - name: example-volume
    persistentVolumeClaim:
      claimName: my-pvc
`

var secretYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: volume-types-pod-secret
  namespace: %s
spec:
  containers:
  - name: volume-types
    image: public.ecr.aws/docker/library/busybox:1.36
  volumes:
  - name: example-volume
    secret:
      secretName: my-secret
`

var projectedYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: volume-types-pod-projected
  namespace: %s
spec:
  containers:
  - name: volume-types
    image: public.ecr.aws/docker/library/busybox:1.36
  volumes:
  - name: example-volume
    projected:
      sources:
      - secret:
          name: my-secret
`

var csiYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: volume-types-pod-csi
  namespace: %s
spec:
  containers:
  - name: volume-types
    image: public.ecr.aws/docker/library/busybox:1.36
  volumes:
  - name: example-volume
    csi:
      driver: example
      volumeAttributes:
        volumeName: example
`

// TEST DATA FOR DEPLOYMENT TESTS

var deploymentYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-%s
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
      - name: volume-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

// TEST DATA FOR REPLICASET TESTS

var rsYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-%s
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
      - name: volume-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

// TEST DATA FOR DAEMONSET TESTS

var containerDSYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types-%s
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
      - name: volume-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

// TEST DATA FOR STATEFULSET TESTS

var containerSSYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-%s
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
      - name: volume-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

// TEST DATA FOR JOB TESTS

var containerJobYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-%s
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: volume-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      restartPolicy: Never
`

// TEST DATA FOR CRONJOB TESTS

var containerCronJobYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-%s
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: volume-types-%s
            image: public.ecr.aws/docker/library/busybox:1.36
          restartPolicy: OnFailure
`

// TEST DATA FOR REPLICATIONCONTROLLER TESTS

var containerRCYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-%s
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
      - name: volume-types-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

// TEST DATA FOR PODTEMPLATE TESTS

var containerPodTemplateYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-%s
  namespace: %s
template:
  spec:
    containers:
    - name: volume-types-%s
      image: public.ecr.aws/docker/library/busybox:1.36
    restartPolicy: Always
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-volume-types": "deny"}

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

	f := features.New("Volume Types tests").
		// POD TESTS
		Assess("Successful deployment of a Pod with a valid configMap volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(configMapYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with a valid downwardAPI volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(downwardAPIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with a valid emptyDir volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(emptyDirYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with a valid ephemeral volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(emphemeralYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with a valid pvc volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(pvcYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with a valid secret volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(secretYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with a valid projected volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(projectedYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with a valid csi volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(csiYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with a prohibited volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, namespace))
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		})
		// DEPLOYMENT TESTS
		// Assess("Successful deployment of a Deployment with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "success", namespace, "success", "100"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).Assess("Successful deployment of a Deployment with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentNoRunAsUserYAML, "success", namespace, "success"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a Deployment with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "rejected", namespace, "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // REPLICASET TESTS
		// Assess("Successful deployment of a ReplicaSet with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "success", namespace, "success", "100"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).Assess("Successful deployment of a ReplicaSet with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSNoRunAsUserYAML, "success", namespace, "success"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicaSet with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "rejected", namespace, "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicaSet as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicaSet with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // DAEMONSET TESTS
		// Assess("Successful deployment of a DaemonSet with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "success", namespace, "success", "100"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).Assess("Successful deployment of a DaemonSet with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSNoRunAsUserYAML, "success", namespace, "success"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a DaemonSet with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "rejected", namespace, "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a DaemonSet as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a DaemonSet with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // STATEFULSET TESTS
		// Assess("Successful deployment of a StatefulSet with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "success", namespace, "success", "100"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).Assess("Successful deployment of a StatefulSet with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSNoRunAsUserYAML, "success", namespace, "success"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a StatefulSet with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "rejected", namespace, "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a StatefulSet as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a StatefulSet with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // JOB TESTS
		// Assess("Successful deployment of a Job with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "success", namespace, "success", "100"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).Assess("Successful deployment of a Job with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobNoRunAsUserYAML, "success", namespace, "success"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a Job with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "rejected", namespace, "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a Job as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a Job with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // CRONJOB TESTS
		// Assess("Successful deployment of a CronJob with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "success", namespace, "success", "100"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).Assess("Successful deployment of a CronJob with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobNoRunAsUserYAML, "success", namespace, "success"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a CronJob with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "rejected", namespace, "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a CronJob as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a CronJob with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // REPLICATIONCONTROLLER TESTS
		// Assess("Successful deployment of a ReplicationController with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "success", namespace, "success", "100"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).Assess("Successful deployment of a ReplicationController with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCNoRunAsUserYAML, "success", namespace, "success"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicationController with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "rejected", namespace, "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicationController as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a ReplicationController with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// // PODTEMPLATE TESTS
		// Assess("Successful deployment of a PodTemplate with container as runAsUser not set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "success", namespace, "success", "100"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).Assess("Successful deployment of a PodTemplate with container as runAsUser is not set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should PASS!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateNoRunAsUserYAML, "success", namespace, "success"))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a PodTemplate with container as container[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "rejected", namespace, "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a PodTemplate as spec.securityContext.runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "rejected", namespace, "0", "rejected", "100"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		// Assess("Rejected deployment of a PodTemplate with initContainer as initContainer[*].runAsUser is set to 0", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	// get namespace
		// 	namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

		// 	// this should FAIL!
		// 	err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "rejected", namespace, "rejected", "100", "rejected", "0"))
		// 	if err == nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// })

		_ = testEnv.Test(t, f.Feature())

	}
