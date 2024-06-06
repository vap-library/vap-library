package pss_running_as_non_root

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
  name: running-as-non-root-%s
  namespace: %s
spec:
  containers:
  - name: running-as-non-root-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: %s
`

var containerWithoutYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: running-as-non-root-%s
  namespace: %s
spec:
  containers:
  - name: running-as-non-root-%s
    image: public.ecr.aws/docker/library/busybox:1.36
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
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: %s
`

var containerOnlyDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: running-as-non-root-only-default-%s
  namespace: %s
spec:
  securityContext:
    runAsNonRoot: %s
  containers:
  - name: running-as-non-root-%s
    image: public.ecr.aws/docker/library/busybox:1.36
`

var twoContainersWithDefaultOnlyOneYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: running-as-non-root-two-with-default-only-one-%s
  namespace: %s
spec:
  securityContext:
    runAsNonRoot: %s
  containers:
  - name: running-as-non-root-two-with-default-only-one-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: %s
  - name: running-as-non-root-two-with-default-only-one-%s
    image: public.ecr.aws/docker/library/busybox:1.36
`

var twoInitContainersWithDefaultOnlyOneYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-running-as-non-root-two-with-default-only-one-%s
  namespace: %s
spec:
  securityContext:
    runAsNonRoot: %s
  containers:
  - name: running-as-non-root-two-with-default-only-one-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: %s
  initContainers:
  - name: init-running-as-non-root-two-with-default-only-one-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: %s
  - name: init-running-as-non-root-two-with-default-only-one-%s
    image: public.ecr.aws/docker/library/busybox:1.36
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
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: %s
  initContainers:
  - name: init-running-as-non-root-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: %s
`

var initContainerWithDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-running-as-non-root-%s
  namespace: %s
spec:
  securityContext:
    runAsNonRoot: %s
  containers:
  - name: running-as-non-root-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: %s
  initContainers:
  - name: init-running-as-non-root-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: %s
`

var initContainerOnlyDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-running-as-non-root-%s
  namespace: %s
spec:
  securityContext:
    runAsNonRoot: %s
  containers:
  - name: running-as-non-root-%s
    image: public.ecr.aws/docker/library/busybox:1.36
  initContainers:
  - name: init-running-as-non-root-%s
    image: public.ecr.aws/docker/library/busybox:1.36
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
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerDeploymentWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerDeploymentOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

var twoContainersDeploymentWithDefaultOnlyOneYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      - name: running-as-non-root-2-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerDeploymentWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerDeploymentOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
`

var twoInitContainersDeploymentWithDefaultOnlyOneYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      - name: init-running-as-non-root-2-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerRSWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerRSOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerRSWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerRSOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerDSWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerDSOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerDSWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerDSOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerSSWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerSSOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerSSWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerSSOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      restartPolicy: Never
`

var containerJobWithDefaultYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: busybox-job-%s
  namespace: %s
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      restartPolicy: Never
`

var containerJobOnlyDefaultYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: busybox-job-%s
  namespace: %s
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      restartPolicy: Never
`

var initContainerJobWithDefaultYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: init-busybox-job-%s
  namespace: %s
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      restartPolicy: Never
`

var initContainerJobOnlyDefaultYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: init-busybox-job-%s
  namespace: %s
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
          - name: running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsNonRoot: %s
          restartPolicy: OnFailure
`

var containerCronJobWithDefaultYAML string = `
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
          securityContext:
            runAsNonRoot: %s
          containers:
          - name: running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsNonRoot: %s
          restartPolicy: OnFailure
`

var containerCronJobOnlyDefaultYAML string = `
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
          securityContext:
            runAsNonRoot: %s
          containers:
          - name: running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
          restartPolicy: OnFailure
`

var twoContainersCronJobWithDefaultOnlyOneYAML string = `
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
          securityContext:
            runAsNonRoot: %s
          containers:
          - name: running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsNonRoot: %s
          - name: running-as-non-root-2-%s
            image: public.ecr.aws/docker/library/busybox:1.36
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
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsNonRoot: %s
          initContainers:
          - name: init-running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsNonRoot: %s
          restartPolicy: OnFailure
`

var initContainerCronJobWithDefaultYAML string = `
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
          securityContext:
            runAsNonRoot: %s
          containers:
          - name: running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsNonRoot: %s
          initContainers:
          - name: init-running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsNonRoot: %s
          restartPolicy: OnFailure
`

var initContainerCronJobOnlyDefaultYAML string = `
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
          securityContext:
            runAsNonRoot: %s
          containers:
          - name: running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
          initContainers:
          - name: init-running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
          restartPolicy: OnFailure
`

var twoInitContainersCronJobWithDefaultOnlyOneYAML string = `
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
          securityContext:
            runAsNonRoot: %s
          containers:
          - name: running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsNonRoot: %s
          initContainers:
          - name: init-running-as-non-root-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              runAsNonRoot: %s
          - name: init-running-as-non-root-2-%s
            image: public.ecr.aws/docker/library/busybox:1.36
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
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerRCWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var containerRCOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerRCWithDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          runAsNonRoot: %s
`

var initContainerRCOnlyDefaultYAML string = `
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
      securityContext:
        runAsNonRoot: %s
      containers:
      - name: running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-running-as-non-root-%s
        image: public.ecr.aws/docker/library/busybox:1.36
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
    - name: running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsNonRoot: %s
    restartPolicy: Always
`

var containerPodTemplateWithDefaultYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: busybox-podtemplate-%s
  namespace: %s
template:
  spec:
    securityContext:
      runAsNonRoot: %s
    containers:
    - name: running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsNonRoot: %s
    restartPolicy: Always
`

var containerPodTemplateOnlyDefaultYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: busybox-podtemplate-%s
  namespace: %s
template:
  spec:
    securityContext:
      runAsNonRoot: %s
    containers:
    - name: running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
    restartPolicy: Always
`

var twoContainersPodTemplateWithDefaultOnlyOneYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: busybox-podtemplate-%s
  namespace: %s
template:
  spec:
    securityContext:
      runAsNonRoot: %s
    containers:
    - name: running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsNonRoot: %s
    - name: running-as-non-root-2-%s
      image: public.ecr.aws/docker/library/busybox:1.36
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
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsNonRoot: %s
    restartPolicy: Always
    initContainers:
    - name: init-running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsNonRoot: %s
`

var initContainerPodTemplateWithDefaultYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: init-busybox-podtemplate-%s
  namespace: %s
template:
  spec:
    securityContext:
      runAsNonRoot: %s
    containers:
    - name: running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsNonRoot: %s
    restartPolicy: Always
    initContainers:
    - name: init-running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsNonRoot: %s
`

var initContainerPodTemplateOnlyDefaultYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: init-busybox-podtemplate-%s
  namespace: %s
template:
  spec:
    securityContext:
      runAsNonRoot: %s
    containers:
    - name: running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
    restartPolicy: Always
    initContainers:
    - name: init-running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
`

var twoInitContainersPodTemplateWithDefaultOnlyOneYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: init-busybox-podtemplate-%s
  namespace: %s
template:
  spec:
    securityContext:
      runAsNonRoot: %s
    containers:
    - name: running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsNonRoot: %s
    restartPolicy: Always
    initContainers:
    - name: init-running-as-non-root-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        runAsNonRoot: %s
    - name: init-running-as-non-root-2-%s
      image: public.ecr.aws/docker/library/busybox:1.36
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
           "runAsNonRoot": %s
         },
         "stdin": true,
         "targetContainerName": "running-as-non-root-ephemeral",
         "terminationMessagePolicy": "File",
         "tty": true
      }
    ]
  }
}
`

var twoContainersEphemeralPatchYAML string = `
{
  "spec": {
    "ephemeralContainers": [
      {
         "image": "public.ecr.aws/docker/library/busybox:1.36",
         "name": "ephemeral",
         "resources": {},
         "securityContext": {
           "runAsNonRoot": %s
         },
         "stdin": true,
         "targetContainerName": "running-as-non-root-ephemeral",
         "terminationMessagePolicy": "File",
         "tty": true
      },
      {
        "image": "public.ecr.aws/docker/library/busybox:1.36",
        "name": "ephemeral-two",
        "resources": {},
        "securityContext": {
          "runAsNonRoot": %s
        },
        "stdin": true,
        "targetContainerName": "running-as-non-root-ephemeral",
        "terminationMessagePolicy": "File",
        "tty": true
     }
    ]
  }
}
`

var twoContainersEphemeralOneUnsetPatchYAML string = `
{
  "spec": {
    "ephemeralContainers": [
      {
         "image": "public.ecr.aws/docker/library/busybox:1.36",
         "name": "ephemeral",
         "resources": {},
         "securityContext": {
           "runAsNonRoot": %s
         },
         "stdin": true,
         "targetContainerName": "running-as-non-root-ephemeral",
         "terminationMessagePolicy": "File",
         "tty": true
      },
      {
        "image": "public.ecr.aws/docker/library/busybox:1.36",
        "name": "ephemeral-two",
        "resources": {},
        "stdin": true,
        "targetContainerName": "running-as-non-root-ephemeral",
        "terminationMessagePolicy": "File",
        "tty": true
     }
    ]
  }
}
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-running-as-non-root": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil)
	if err != nil {
		log.Fatalf("Unable to create Kind cluster for test. Error msg: %s", err)
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestRunAsNonRoot(t *testing.T) {

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
		Assess("Successful deployment of a Pod with two containers as spec.runAsNonRoot set to true and container.runAsNonRoot is set to true for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoContainersWithDefaultOnlyOneYAML, "success", namespace, "true", "success-01", "true", "success-02"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "success", namespace, "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerWithDefaultYAML, "success-default-true", namespace, "true", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerWithDefaultYAML, "success-default-false", namespace, "false", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerOnlyDefaultYAML, "success-only-default-false", namespace, "true", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with two initContainers as spec.runAsNonRoot set to true and container.runAsNonRoot is set to true for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoInitContainersWithDefaultOnlyOneYAML, "success", namespace, "true", "success-01", "true", "success-02", "true", "success-03"))
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
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod without runAsNonRoot is set", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithoutYAML, "rejected", namespace, "rejected"))
			if err == nil {
				t.Fatal("containers without any of the runAsNonRoot fields set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal("containers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "rejected", namespace, "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		// // DEPLOYMENT TESTS
		Assess("Successful deployment of a Deployment with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "success", namespace, "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentWithDefaultYAML, "success-default-true", namespace, "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentWithDefaultYAML, "success-default-false", namespace, "false", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with container as container.runAsNonRoot is not defined and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentOnlyDefaultYAML, "success-only-default-true", namespace, "true", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with two containers as spec.runAsNonRoot set to true and container.runAsNonRoot is set to true for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoContainersDeploymentWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "true", "success-01", "true", "success-02"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "success", namespace, "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentWithDefaultYAML, "success-default-true", namespace, "true", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentWithDefaultYAML, "success-default-false", namespace, "false", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentOnlyDefaultYAML, "success-only-default-false", namespace, "true", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with two initContainers as spec.runAsNonRoot set to true and container.runAsNonRoot is set to true for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoInitContainersDeploymentWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "true", "success-01", "true", "success-02", "true", "success-03"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "rejected", namespace, "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal("containers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "rejected", namespace, "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		// // REPLICASET TESTS
		Assess("Successful deployment of a ReplicaSet with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "success", namespace, "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "success-default-true", namespace, "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "success-default-false", namespace, "false", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with container as container.runAsNonRoot is not defined and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSOnlyDefaultYAML, "success-only-default-true", namespace, "true", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "success", namespace, "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSWithDefaultYAML, "success-default-true", namespace, "true", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSWithDefaultYAML, "success-default-false", namespace, "false", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSOnlyDefaultYAML, "success-only-default-false", namespace, "true", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "rejected", namespace, "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal("containers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "rejected", namespace, "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		// // DAEMONSET TESTS
		Assess("Successful deployment of a DaemonSet with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "success", namespace, "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "success-default-true", namespace, "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "success-default-false", namespace, "false", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with container as container.runAsNonRoot is not defined and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSOnlyDefaultYAML, "success-only-default-true", namespace, "true", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "success", namespace, "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSWithDefaultYAML, "success-default-true", namespace, "true", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSWithDefaultYAML, "success-default-false", namespace, "false", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSOnlyDefaultYAML, "success-only-default-false", namespace, "true", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "rejected", namespace, "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal("containers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "rejected", namespace, "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		// // STATEFULSET TESTS
		Assess("Successful deployment of a Stateful Set with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "success", namespace, "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "success-default-true", namespace, "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "success-default-false", namespace, "false", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with container as container.runAsNonRoot is not defined and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSOnlyDefaultYAML, "success-only-default-true", namespace, "true", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "success", namespace, "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSWithDefaultYAML, "success-default-true", namespace, "true", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSWithDefaultYAML, "success-default-false", namespace, "false", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSOnlyDefaultYAML, "success-only-default-false", namespace, "true", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "rejected", namespace, "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal("containers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "rejected", namespace, "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		// // JOB TESTS
		Assess("Successful deployment of a Job with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "success", namespace, "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "success-default-true", namespace, "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "success-default-false", namespace, "false", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with container as container.runAsNonRoot is not defined and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobOnlyDefaultYAML, "success-only-default-true", namespace, "true", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "success", namespace, "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobWithDefaultYAML, "success-default-true", namespace, "true", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobWithDefaultYAML, "success-default-false", namespace, "false", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobOnlyDefaultYAML, "success-only-default-false", namespace, "true", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "rejected", namespace, "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal("containers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "rejected", namespace, "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		// // CRONJOB TESTS
		Assess("Successful deployment of a CronJob with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "success", namespace, "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "success-default-true", namespace, "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "success-default-false", namespace, "false", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with container as container.runAsNonRoot is not defined and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobOnlyDefaultYAML, "success-only-default-true", namespace, "true", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with two containers as spec.runAsNonRoot set to true and container.runAsNonRoot is set to true for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoContainersCronJobWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "true", "success-01", "true", "success-02"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "success", namespace, "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobWithDefaultYAML, "success-default-true", namespace, "true", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobWithDefaultYAML, "success-default-false", namespace, "false", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobOnlyDefaultYAML, "success-only-default-false", namespace, "true", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with two initContainers as spec.runAsNonRoot set to true and container.runAsNonRoot is set to true for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoInitContainersCronJobWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "true", "success-01", "true", "success-02", "true", "success-03"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "rejected", namespace, "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal("containers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "rejected", namespace, "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		// // REPLICATIONCONTROLLER TESTS
		Assess("Successful deployment of a ReplicationController with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "success", namespace, "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "success-default-true", namespace, "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "success-default-false", namespace, "false", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with container as container.runAsNonRoot is not defined and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCOnlyDefaultYAML, "success-only-default-true", namespace, "true", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "success", namespace, "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCWithDefaultYAML, "success-default-true", namespace, "true", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCWithDefaultYAML, "success-default-false", namespace, "false", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCOnlyDefaultYAML, "success-only-default-false", namespace, "true", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "rejected", namespace, "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal("containers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "rejected", namespace, "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		// // PODTEMPLATE TESTS
		Assess("Successful deployment of a PodTemplate with container as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "success", namespace, "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "success-default-true", namespace, "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with container as container.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "success-default-false", namespace, "false", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with container as container.runAsNonRoot is not defined and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateOnlyDefaultYAML, "success-only-default-true", namespace, "true", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with two containers as spec.runAsNonRoot set to true and container.runAsNonRoot is set to true for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoContainersPodTemplateWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "true", "success-01", "true", "success-02"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as runAsNonRoot is set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "success", namespace, "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateWithDefaultYAML, "success-default-true", namespace, "true", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as initContainer.runAsNonRoot is set to true and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateWithDefaultYAML, "success-default-false", namespace, "false", "success", "true", "success", "true"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateOnlyDefaultYAML, "success-only-default-false", namespace, "true", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with two initContainers as spec.runAsNonRoot set to true and container.runAsNonRoot is set to true for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoInitContainersPodTemplateWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "true", "success-01", "true", "success-02", "true", "success-03"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "rejected", namespace, "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "false"))
			if err == nil {
				t.Fatal("containers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as container.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "false"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as container.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as runAsNonRoot is set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "rejected", namespace, "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to true", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateWithDefaultYAML, "rejected-default-true", namespace, "true", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers with securityContext.runAsNonRoot field set to false were accepted, when the lower-priority spec.securityContext.runAsRoot was set to true")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as initContainer.runAsNonRoot is set to false, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateWithDefaultYAML, "rejected-default-false", namespace, "false", "rejected", "true", "rejected", "false"))
			if err == nil {
				t.Fatal("initContainers without securityContext.runAsNonRoot field set to true were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as initContainer.runAsNonRoot is not defined, and spec.runAsNonRoot set to false", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateOnlyDefaultYAML, "rejected-only-default-false", namespace, "false", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.runAsNonRoot field set to true were accepted")
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

			// create a pod that will be used for certain ephemeral container tests
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "ephemeral", namespace, "ephemeral", "true"))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the pod
			time.Sleep(2 * time.Second)

			// create a pod with spec.securityContext.runAsNonRoot that will be used for certain ephemeral container tests
			err = testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerOnlyDefaultYAML, "ephemeral", namespace, "true", "ephemeral"))
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
			err = client.Resources(namespace).Get(ctx, "running-as-non-root-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "false"))

			patch := k8s.Patch{patchType, patchData}

			// patch the pod, this should FAIL!
			err = client.Resources(namespace).PatchSubresource(ctx, pod, "ephemeralcontainers", patch)
			if err == nil {
				t.Fatal("ephemeral container without securityContext.runAsNonRoot field should be rejected")
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
			err = client.Resources(namespace).Get(ctx, "running-as-non-root-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "true"))

			patch := k8s.Patch{patchType, patchData}

			// patch the pod
			err = client.Resources(namespace).PatchSubresource(ctx, pod, "ephemeralcontainers", patch)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Two valid ephemeral containers are accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// get client
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			// get the pod that was created in setup to attach an ephemeral container to it
			pod := &v1.Pod{}
			err = client.Resources(namespace).Get(ctx, "running-as-non-root-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(twoContainersEphemeralPatchYAML, "true", "true"))

			patch := k8s.Patch{patchType, patchData}

			// patch the pod
			err = client.Resources(namespace).PatchSubresource(ctx, pod, "ephemeralcontainers", patch)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Two ephemeral containers are accepted, where spec.runAsNonRoot set to true and container.runAsNonRoot is set to true for one ephemeral container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// get client
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			// get the pod that was created in setup to attach an ephemeral container to it
			pod := &v1.Pod{}
			err = client.Resources(namespace).Get(ctx, "running-as-non-root-only-default-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(twoContainersEphemeralOneUnsetPatchYAML, "true"))

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
