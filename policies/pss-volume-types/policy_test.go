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

// TEST DATA FOR POD TESTS

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

// TEST DATA FOR DEPLOYMENT TESTS

var deploymentYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-invalid
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        hostPath:
          path: /data/foo # directory location on host
          type: Directory # this field is optional
`

var deploymentConfigMapYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-configmap
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

var deploymentCSIYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-csi
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        csi:
          driver: example
          volumeAttributes:
            volumeName: example
`

var deploymentDownwardAPIYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-downwardapi
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

var deploymentEmphemeralYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-ephemeral
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

var deploymentEmptyDirYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-emptydir
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        emptyDir:
          sizeLimit: 500Mi
`

var deploymentProjectedYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-projected
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        projected:
          sources:
          - secret:
              name: my-secret
`

var deploymentPVCYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-pvc
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        persistentVolumeClaim:
          claimName: my-pvc
`

var deploymentSecretYAML string = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-volume-types-secret
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        secret:
          secretName: my-secret
`

// TEST DATA FOR REPLICASET TESTS

var rsYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-invalid
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        hostPath:
          path: /data/foo # directory location on host
          type: Directory # this field is optional
`

var rsConfigMapYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-configmap
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

var rsCSIYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-csi
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        csi:
          driver: example
          volumeAttributes:
            volumeName: example
`

var rsDownwardAPIYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-downwardapi
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

var rsEphemeralYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-ephemeral
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

var rsEmptyDirYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-emptydir
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        emptyDir:
          sizeLimit: 500Mi
`

var rsProjectedYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-projected
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        projected:
          sources:
          - secret:
              name: my-secret
`

var rsPVCYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-pvc
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        persistentVolumeClaim:
          claimName: my-pvc
`

var rsSecretYAML string = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replicaset-volume-types-secret
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        secret:
          secretName: my-secret
`

// TEST DATA FOR DAEMONSET TESTS

var dsYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        hostPath:
          path: /data/foo # directory location on host
          type: Directory # this field is optional
`

var dsConfigMapYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types
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

var dsCSIYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types-csi
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
        - name: example-volume
          csi:
            driver: example
            volumeAttributes:
              volumeName: example
`

var dsDownwardAPIYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types-downwardapi
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

var dsEphemeralYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types-ephemeral
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

var dsEmptyDirYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types-emptydir
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        emptyDir:
          sizeLimit: 500Mi
`

var dsProjectedYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types-projected
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        projected:
          sources:
          - secret:
              name: my-secret
`

var dsPVCYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types-pvc
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        persistentVolumeClaim:
          claimName: my-pvc
`

var dsSecretYAML string = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: daemonset-volume-types-secret
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        secret:
          secretName: my-secret
`

// TEST DATA FOR STATEFULSET TESTS

var ssYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-invalid
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        hostPath:
          path: /data/foo # directory location on host
          type: Directory # this field is optional
`

var ssConfigMapYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-configmap
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

var ssCSIYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-csi
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
        - name: example-volume
          csi:
            driver: example
            volumeAttributes:
              volumeName: example
`

var ssDownwardAPIYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-downwardapi
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

var ssEphemeralYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-ephemeral
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

var ssEmptyDirYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-emptydir
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        emptyDir:
          sizeLimit: 500Mi
`

var ssProjectedYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-projected
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        projected:
          sources:
          - secret:
              name: my-secret
`

var ssPVCYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-pvc
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        persistentVolumeClaim:
          claimName: my-pvc
`

var ssSecretYAML string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: statefulset-volume-types-secret
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        secret:
          secretName: my-secret
`

// TEST DATA FOR JOB TESTS

var jobYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-invalid
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        hostPath:
          path: /data/foo # directory location on host
          type: Directory # this field is optional
      restartPolicy: Never
`

var jobConfigMapYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-configmap
  namespace: %s
spec:
  template:
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
      restartPolicy: Never
`

var jobCSIYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-csi
  namespace: %s
spec:
  template:
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
      restartPolicy: Never
`

var jobDownwardAPIYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-downwardapi
  namespace: %s
spec:
  template:
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
      restartPolicy: Never
`

var jobEphemeralYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-ephemeral
  namespace: %s
spec:
  template:
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
      restartPolicy: Never
`

var jobEmptyDirYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-emptydir
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        emptyDir:
          sizeLimit: 500Mi
      restartPolicy: Never
`

var jobProjectedYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-projected
  namespace: %s
spec:
  template:
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
      restartPolicy: Never
`

var jobPVCYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-pvc
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        persistentVolumeClaim:
          claimName: my-pvc
      restartPolicy: Never
`

var jobSecretYAML string = `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-volume-types-secret
  namespace: %s
spec:
  template:
    spec:
      containers:
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        secret:
          secretName: my-secret
      restartPolicy: Never
`

// TEST DATA FOR CRONJOB TESTS

var cronJobYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-invalid
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: volume-types
            image: public.ecr.aws/docker/library/busybox:1.36
          volumes:
          - name: example-volume
            hostPath:
             path: /data/foo # directory location on host
             type: Directory # this field is optional
          restartPolicy: OnFailure
          
`

var cronJobConfigMapYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-configmap
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
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
          restartPolicy: OnFailure
`

var cronJobCSIYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-csi
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
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
          restartPolicy: OnFailure
`

var cronJobDownwardAPIYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-downwardapi
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
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
          restartPolicy: OnFailure
`

var cronJobEphemeralYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-ephemeral
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
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
          restartPolicy: OnFailure
`

var cronJobEmptyDirYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-emptydir
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: volume-types
            image: public.ecr.aws/docker/library/busybox:1.36
          volumes:
          - name: example-volume
            emptyDir:
              sizeLimit: 500Mi
          restartPolicy: OnFailure
`

var cronJobProjectedYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-projected
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
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
          restartPolicy: OnFailure
`

var cronJobPVCYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-pvc
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: volume-types
            image: public.ecr.aws/docker/library/busybox:1.36
          volumes:
          - name: example-volume
            persistentVolumeClaim:
              claimName: my-pvc
          restartPolicy: OnFailure
`

var cronJobSecretYAML string = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cronjob-volume-types-secret
  namespace: %s
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: volume-types
            image: public.ecr.aws/docker/library/busybox:1.36
          volumes:
          - name: example-volume
            secret:
              secretName: my-secret
          restartPolicy: OnFailure
`

// TEST DATA FOR REPLICATIONCONTROLLER TESTS

var rcYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-invalid
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        hostPath:
          path: /data/foo # directory location on host
          type: Directory # this field is optional
`

var rcConfigMapYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-configmap
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

var rcCSIYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-csi
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        csi:
          driver: example
          volumeAttributes:
            volumeName: example
`

var rcDownwardAPIYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-downwardapi
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

var rcEphemeralYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-ephemeral
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

var rcEmptyDirYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-emptydir
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        emptyDir:
          sizeLimit: 500Mi
`

var rcProjectedYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-projected
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        projected:
          sources:
          - secret:
              name: my-secret
`

var rcPVCYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-pvc
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        persistentVolumeClaim:
          claimName: my-pvc
`

var rcSecretYAML string = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replicationcontroller-volume-types-secret
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
      - name: volume-types
        image: public.ecr.aws/docker/library/busybox:1.36
      volumes:
      - name: example-volume
        secret:
          secretName: my-secret
`

// TEST DATA FOR PODTEMPLATE TESTS

var podTemplateYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-invalid
  namespace: %s
template:
  spec:
    containers:
    - name: volume-types
      image: public.ecr.aws/docker/library/busybox:1.36
    volumes:
    - name: example-volume
      hostPath:
        path: /data/foo # directory location on host
        type: Directory # this field is optional
    restartPolicy: Always
`

var podTemplateConfigMapYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-configmap
  namespace: %s
template:
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
    restartPolicy: Always
`

var podTemplateCSIYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-csi
  namespace: %s
template:
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
    restartPolicy: Always
`

var podTemplateDownwardAPIYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-downwardapi
  namespace: %s
template:
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
    restartPolicy: Always
`

var podTemplateEphemeralYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-ephemeral
  namespace: %s
template:
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
    restartPolicy: Always
`

var podTemplateEmptyDirYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-emptydir
  namespace: %s
template:
  spec:
    containers:
    - name: volume-types
      image: public.ecr.aws/docker/library/busybox:1.36
    volumes:
    - name: example-volume
      emptyDir:
        sizeLimit: 500Mi
    restartPolicy: Always
`

var podTemplateProjectedYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-projected
  namespace: %s
template:
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
    restartPolicy: Always
`

var podTemplatePVCYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-pvc
  namespace: %s
template:
  spec:
    containers:
    - name: volume-types
      image: public.ecr.aws/docker/library/busybox:1.36
    volumes:
    - name: example-volume
      persistentVolumeClaim:
        claimName: my-pvc
    restartPolicy: Always
`

var podTemplateSecretYAML string = `
apiVersion: v1
kind: PodTemplate
metadata:
  name: podtemplate-volume-types-secret
  namespace: %s
template:
  spec:
    containers:
    - name: volume-types
      image: public.ecr.aws/docker/library/busybox:1.36
    volumes:
    - name: example-volume
      secret:
        secretName: my-secret
    restartPolicy: Always
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/pss-volume-types": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil)
	if err != nil {
		log.Fatalf("Unable to create Kind cluster for test. Error msg: %s", err)
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestVolumeTypes(t *testing.T) {

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
				t.Fatal("pods with a prohibited volume type were accepted")
			}

			return ctx
		}).
		// DEPLOYMENT TESTS
		Assess("Successful deployment of a Deployment with a valid configMap volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(deploymentConfigMapYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with a valid downwardAPI volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(deploymentDownwardAPIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with a valid emptyDir volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(deploymentEmptyDirYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with a valid ephemeral volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(deploymentEmphemeralYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with a valid pvc volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(deploymentPVCYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with a valid secret volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(deploymentSecretYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with a valid projected volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(deploymentProjectedYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Deployment with a valid csi volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(deploymentCSIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Deployment with a prohibited volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(deploymentYAML, namespace))
			if err == nil {
				t.Fatal("pods with a prohibited volume type were accepted")
			}

			return ctx
		}).
		// // REPLICASET TESTS
		Assess("Successful deployment of a ReplicaSet with a valid configMap volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rsConfigMapYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with a valid downwardAPI volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rsDownwardAPIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with a valid emptyDir volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rsEmptyDirYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with a valid ephemeral volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rsEphemeralYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with a valid pvc volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rsPVCYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with a valid secret volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rsSecretYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with a valid projected volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rsProjectedYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicaSet with a valid csi volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rsCSIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicaSet with a prohibited volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rsYAML, namespace))
			if err == nil {
				t.Fatal("pods with a prohibited volume type were accepted")
			}

			return ctx
		}).
		// // DAEMONSET TESTS
		Assess("Successful deployment of a DaemonSet with a valid configMap volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dsConfigMapYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with a valid downwardAPI volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dsDownwardAPIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with a valid emptyDir volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dsEmptyDirYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with a valid ephemeral volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dsEphemeralYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with a valid pvc volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dsPVCYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with a valid secret volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dsSecretYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with a valid projected volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dsProjectedYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a DaemonSet with a valid csi volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dsCSIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a DaemonSet with a prohibited volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(dsYAML, namespace))
			if err == nil {
				t.Fatal("pods with a prohibited volume type were accepted")
			}

			return ctx
		}).
		// // STATEFULSET TESTS
		Assess("Successful deployment of a StatefulSet with a valid configMap volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ssConfigMapYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with a valid downwardAPI volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ssDownwardAPIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with a valid emptyDir volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ssEmptyDirYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with a valid ephemeral volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ssEphemeralYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with a valid pvc volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ssPVCYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with a valid secret volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ssSecretYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with a valid projected volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ssProjectedYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a StatefulSet with a valid csi volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ssCSIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a StatefulSet with a prohibited volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(ssYAML, namespace))
			if err == nil {
				t.Fatal("pods with a prohibited volume type were accepted")
			}

			return ctx
		}).
		// // JOB TESTS
		Assess("Successful deployment of a Job with a valid configMap volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(jobConfigMapYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with a valid downwardAPI volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(jobDownwardAPIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with a valid emptyDir volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(jobEmptyDirYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with a valid ephemeral volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(jobEphemeralYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with a valid pvc volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(jobPVCYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with a valid secret volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(jobSecretYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with a valid projected volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(jobProjectedYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a Job with a valid csi volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(jobCSIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a Job with a prohibited volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(jobYAML, namespace))
			if err == nil {
				t.Fatal("pods with a prohibited volume type were accepted")
			}

			return ctx
		}).
		// // CRONJOB TESTS
		Assess("Successful deployment of a CronJob with a valid configMap volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(cronJobConfigMapYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with a valid downwardAPI volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(cronJobDownwardAPIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with a valid emptyDir volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(cronJobEmptyDirYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with a valid ephemeral volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(cronJobEphemeralYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with a valid pvc volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(cronJobPVCYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with a valid secret volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(cronJobSecretYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with a valid projected volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(cronJobProjectedYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a CronJob with a valid csi volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(cronJobCSIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a CronJob with a prohibited volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(cronJobYAML, namespace))
			if err == nil {
				t.Fatal("pods with a prohibited volume type were accepted")
			}

			return ctx
		}).
		// REPLICATIONCONTROLLER TESTS
		Assess("Successful deployment of a ReplicationController with a valid configMap volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rcConfigMapYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with a valid downwardAPI volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rcDownwardAPIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with a valid emptyDir volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rcEmptyDirYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with a valid ephemeral volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rcEphemeralYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with a valid pvc volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rcPVCYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with a valid secret volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rcSecretYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with a valid projected volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rcProjectedYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a ReplicationController with a valid csi volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rcCSIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a ReplicationController with a prohibited volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(rcYAML, namespace))
			if err == nil {
				t.Fatal("pods with a prohibited volume type were accepted")
			}

			return ctx
		}).
		// PODTEMPLATE TESTS
		Assess("Successful deployment of a PodTemplate with a valid configMap volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(podTemplateConfigMapYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with a valid downwardAPI volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(podTemplateDownwardAPIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with a valid emptyDir volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(podTemplateEmptyDirYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with a valid ephemeral volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(podTemplateEphemeralYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with a valid pvc volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(podTemplatePVCYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with a valid secret volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(podTemplateSecretYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with a valid projected volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(podTemplateProjectedYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Successful deployment of a PodTemplate with a valid csi volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(podTemplateCSIYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Rejected deployment of a PodTemplate with a prohibited volume configuration", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(podTemplateYAML, namespace))
			if err == nil {
				t.Fatal("pods with a prohibited volume type were accepted")
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
