package pss_seccomp

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
  name: seccomp-%s
  namespace: %s
spec:
  containers:
  - name: seccomp-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      seccompProfile:
        type: %s
`

var containerWithDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: seccomp-with-default-%s
  namespace: %s
spec:
  securityContext:
    seccompProfile:
      type: %s
  containers:
  - name: seccomp-with-default-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      seccompProfile:
        type: %s
`

var containerOnlyDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: seccomp-%s
  namespace: %s
spec:
  securityContext:
    seccompProfile:
      type: %s
  containers:
  - name: seccomp-%s
    image: public.ecr.aws/docker/library/busybox:1.36
`

var containerOnlyDefaultWithOtherSecurityContextYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: seccomp-%s
  namespace: %s
spec:
  securityContext:
    seccompProfile:
      type: %s
  containers:
  - name: seccomp-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      runAsNonRoot: false
`

var twoContainersWithDefaultOnlyOneYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: seccomp-two-with-default-only-one-%s
  namespace: %s
spec:
  securityContext:
    seccompProfile:
      type: %s
  containers:
  - name: seccomp-two-with-default-only-one-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      seccompProfile:
        type: %s
  - name: seccomp-two-with-default-only-one-%s
    image: public.ecr.aws/docker/library/busybox:1.36
`

var initContainerYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-seccomp-%s
  namespace: %s
spec:
  containers:
  - name: seccomp-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      seccompProfile:
        type: %s
  initContainers:
  - name: init-seccomp-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      seccompProfile:
        type: %s
`

var initContainerWithDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-seccomp-%s
  namespace: %s
spec:
  securityContext:
    seccompProfile:
      type: %s
  containers:
  - name: seccomp-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      seccompProfile:
        type: %s
  initContainers:
  - name: init-seccomp-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      seccompProfile:
        type: %s
`

var initContainerOnlyDefaultYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-seccomp-%s
  namespace: %s
spec:
  securityContext:
    seccompProfile:
      type: %s
  containers:
  - name: seccomp-%s
    image: public.ecr.aws/docker/library/busybox:1.36
  initContainers:
  - name: init-seccomp-%s
    image: public.ecr.aws/docker/library/busybox:1.36
`

var twoInitContainersWithDefaultOnlyOneYAML string = `
apiVersion: v1
kind: Pod
metadata:
  name: init-seccomp-two-with-default-only-one-%s
  namespace: %s
spec:
  securityContext:
    seccompProfile:
      type: %s
  containers:
  - name: seccomp-two-with-default-only-one-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      seccompProfile:
        type: %s
  initContainers:
  - name: init-seccomp-two-with-default-only-one-%s
    image: public.ecr.aws/docker/library/busybox:1.36
    securityContext:
      seccompProfile:
        type: %s
  - name: init-seccomp-two-with-default-only-one-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      - name: seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-seccomp-%s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      - name: init-seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-seccomp-%s
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
          - name: seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              seccompProfile:
                type: %s
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
            seccompProfile:
              type: %s
          containers:
          - name: seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              seccompProfile:
                type: %s
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
            seccompProfile:
              type: %s
          containers:
          - name: seccomp-%s
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
            seccompProfile:
              type: %s
          containers:
          - name: seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              seccompProfile:
                type: %s
          - name: seccomp-%s
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
          - name: seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              seccompProfile:
                type: %s
          initContainers:
          - name: init-seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              seccompProfile:
                 type: %s
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
            seccompProfile:
              type: %s
          containers:
          - name: seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              seccompProfile:
                type: %s
          initContainers:
          - name: init-seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              seccompProfile:
                type: %s
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
            seccompProfile:
              type: %s
          containers:
          - name: seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
          initContainers:
          - name: init-seccomp-%s
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
            seccompProfile:
              type: %s
          containers:
          - name: seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              seccompProfile:
                type: %s
          initContainers:
          - name: init-seccomp-%s
            image: public.ecr.aws/docker/library/busybox:1.36
            securityContext:
              seccompProfile:
                type: %s
          - name: init-seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
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
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
      initContainers:
      - name: init-seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
        securityContext:
          seccompProfile:
            type: %s
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
        seccompProfile:
          type: %s
      containers:
      - name: seccomp-%s
        image: public.ecr.aws/docker/library/busybox:1.36
      initContainers:
      - name: init-seccomp-%s
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
    - name: seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        seccompProfile:
          type: %s
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
      seccompProfile:
          type: %s
    containers:
    - name: seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        seccompProfile:
          type: %s
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
      seccompProfile:
        type: %s
    containers:
    - name: seccomp-%s
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
      seccompProfile:
          type: %s
    containers:
    - name: seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        seccompProfile:
          type: %s
    - name: seccomp-%s
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
    - name: seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        seccompProfile:
          type: %s
    restartPolicy: Always
    initContainers:
    - name: init-seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        seccompProfile:
          type: %s
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
      seccompProfile:
        type: %s
    containers:
    - name: seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        seccompProfile:
          type: %s
    restartPolicy: Always
    initContainers:
    - name: init-seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        seccompProfile:
          type: %s
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
      seccompProfile:
        type: %s
    containers:
    - name: seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
    restartPolicy: Always
    initContainers:
    - name: init-seccomp-%s
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
      seccompProfile:
        type: %s
    containers:
    - name: seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        seccompProfile:
          type: %s
    restartPolicy: Always
    initContainers:
    - name: init-seccomp-%s
      image: public.ecr.aws/docker/library/busybox:1.36
      securityContext:
        seccompProfile:
          type: %s
    - name: init-seccomp-%s
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
           "seccompProfile": {
             "type": "%s"
           } 
         },
         "stdin": true,
         "targetContainerName": "seccomp-ephemeral",
         "terminationMessagePolicy": "File",
         "tty": true
      }
    ]
  }
}
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-seccomp": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil)
	if err != nil {
		log.Fatalf("Unable to create Kind cluster for test. Error msg: %s", err)
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestSeccomp(t *testing.T) {

	f := features.New("Seccomp tests").
		// POD TESTS
		Assess("Successful deployment of a Pod with container as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "success", namespace, "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with container as container.seccompProfile.type is not defined and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with container as container.seccompProfile.type is not defined, container.securityContext is defined for something else and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerOnlyDefaultWithOtherSecurityContextYAML, "success-only-default-with-other-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with two containers as spec.seccompProfile.type set to RuntimeDefault and container.seccompProfile.type is set to RuntimeDefault for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoContainersWithDefaultOnlyOneYAML, "success", namespace, "RuntimeDefault", "success-01", "RuntimeDefault", "success-02"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with initContainer as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "success", namespace, "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Pod with two initContainers as spec.seccompProfile.type set to RuntimeDefault and container.seccompProfile.type is set to RuntimeDefault for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoInitContainersWithDefaultOnlyOneYAML, "success", namespace, "RuntimeDefault", "success-01", "RuntimeDefault", "success-02", "RuntimeDefault", "success-03"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "rejected", namespace, "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to something else then the allowed values", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "SomethingElseThanAllowed", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerYAML, "rejected", namespace, "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected", "rejected"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Pod with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerWithDefaultYAML, "success-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		// DEPLOYMENT TESTS
		Assess("Successful deployment of a Deployment with container as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "success", namespace, "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with container as container.seccompProfile.type is not defined and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with two containers as spec.seccompProfile.type set to RuntimeDefault and container.seccompProfile.type is set to RuntimeDefault for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoContainersDeploymentWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "RuntimeDefault", "success-01", "RuntimeDefault", "success-02"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "success", namespace, "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with two initContainers as spec.seccompProfile.type set to RuntimeDefault and container.seccompProfile.type is set to RuntimeDefault for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoInitContainersDeploymentWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "RuntimeDefault", "success-01", "RuntimeDefault", "success-02", "RuntimeDefault", "success-03"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentYAML, "rejected", namespace, "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDeploymentWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentYAML, "rejected", namespace, "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDeploymentWithDefaultYAML, "success-default-unconfined", namespace, "Unconfined", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		// REPLICASET TESTS
		Assess("Successful deployment of a ReplicaSet with container as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "success", namespace, "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with container as container.seccompProfile.type is not defined and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "success", namespace, "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSYAML, "rejected", namespace, "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSYAML, "rejected", namespace, "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		// DAEMONSET TESTS
		Assess("Successful deployment of a DaemonSet with container as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "success", namespace, "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with container as container.seccompProfile.type is not defined and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "success", namespace, "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSYAML, "rejected", namespace, "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerDSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSYAML, "rejected", namespace, "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerDSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		// STATEFULSET TESTS
		Assess("Successful deployment of a Stateful Set with container as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "success", namespace, "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with container as container.seccompProfile.type is not defined and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with initContainer as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "success", namespace, "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Stateful Set with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with container as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSYAML, "rejected", namespace, "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerSSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with initContainer as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSYAML, "rejected", namespace, "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Stateful Set with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerSSWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		// JOB TESTS
		Assess("Successful deployment of a Job with container as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "success", namespace, "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with container as container.seccompProfile.type is not defined and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "success", namespace, "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobYAML, "rejected", namespace, "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerJobWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobYAML, "rejected", namespace, "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerJobWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		// CRONJOB TESTS
		Assess("Successful deployment of a CronJob with container as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "success", namespace, "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with container as container.seccompProfile.type is not defined and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with two containers as spec.seccompProfile.type set to RuntimeDefault and container.seccompProfile.type is set to RuntimeDefault for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoContainersCronJobWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "RuntimeDefault", "success-01", "RuntimeDefault", "success-02"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "success", namespace, "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with two initContainers as spec.runAsNonRoot set to RuntimeDefault and container.runAsNonRoot is set to RuntimeDefault for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoInitContainersCronJobWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "RuntimeDefault", "success-01", "RuntimeDefault", "success-02", "RuntimeDefault", "success-03"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobYAML, "rejected", namespace, "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerCronJobWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobYAML, "rejected", namespace, "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerCronJobWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		// REPLICATIONCONTROLLER TESTS
		Assess("Successful deployment of a ReplicationController with container as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "success", namespace, "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with container as container.seccompProfile.type is not defined and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "success", namespace, "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCYAML, "rejected", namespace, "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerRCWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCYAML, "rejected", namespace, "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerRCWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		// PODTEMPLATE TESTS
		Assess("Successful deployment of a PodTemplate with container as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "success", namespace, "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with container as container.seccompProfile.type is not defined and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with two containers as spec.seccompProfile.type set to RuntimeDefault and container.seccompProfile.type is set to RuntimeDefault for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoContainersPodTemplateWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "RuntimeDefault", "success-01", "RuntimeDefault", "success-02"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as seccompProfile.type is set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "success", namespace, "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateWithDefaultYAML, "success-default-rd", namespace, "RuntimeDefault", "success", "RuntimeDefault", "success", "RuntimeDefault"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateOnlyDefaultYAML, "success-only-default-rd", namespace, "RuntimeDefault", "success", "success"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with two initContainers as spec.seccompProfile.type set to RuntimeDefault and container.seccompProfile.type is set to RuntimeDefault for one container and unset for the other", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(twoInitContainersPodTemplateWithDefaultOnlyOneYAML, "success-two-container-default", namespace, "RuntimeDefault", "success-01", "RuntimeDefault", "success-02", "RuntimeDefault", "success-03"))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateYAML, "rejected", namespace, "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as container.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as container.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with container as container.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerPodTemplateWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as seccompProfile.type is set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateYAML, "rejected", namespace, "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to RuntimeDefault or Localhost", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateWithDefaultYAML, "rejected-default-rd", namespace, "RuntimeDefault", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers with securityContext.seccompProfile.type field set to Unconfined were accepted, when the lower-priority spec.securityContext.seccompProfile.type was set to RuntimeDefault or Localhost")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as initContainer.seccompProfile.type is set to Unconfined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "Unconfined"))
			if err == nil {
				t.Fatal("initContainers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as initContainer.seccompProfile.type is not defined, and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateOnlyDefaultYAML, "rejected-only-default-unconfined", namespace, "Unconfined", "rejected", "rejected"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with initContainer as initContainer.seccompProfile.type is set to RuntimeDefault or Localhost and spec.seccompProfile.type set to Unconfined", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(initContainerPodTemplateWithDefaultYAML, "rejected-default-unconfined", namespace, "Unconfined", "rejected", "RuntimeDefault", "rejected", "RuntimeDefault"))
			if err == nil {
				t.Fatal("containers without securityContext.seccompProfile.type field set to RuntimeDefault or Localhost were accepted")
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
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(containerYAML, "ephemeral", namespace, "ephemeral", "RuntimeDefault"))
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
			err = client.Resources(namespace).Get(ctx, "seccomp-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "Unconfined"))

			patch := k8s.Patch{patchType, patchData}

			// patch the pod, this should FAIL!
			err = client.Resources(namespace).PatchSubresource(ctx, pod, "ephemeralcontainers", patch)
			if err == nil {
				t.Fatal("ephemeral container without securityContext.seccompProfile.type field should be rejected")
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
			err = client.Resources(namespace).Get(ctx, "seccomp-ephemeral", namespace, pod)
			if err != nil {
				t.Fatal(err)
			}

			// define patch type
			patchType := types.StrategicMergePatchType

			// define patch data
			patchData := []byte(fmt.Sprintf(containerEphemeralPatchYAML, "RuntimeDefault"))

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
