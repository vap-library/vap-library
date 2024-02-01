# Description
This Validating Admission Policy ensures that configmaps (or secrets) that are defining grafana dashboards are using
their own namespace for folders in grafana.

# Pre-requirements
Grafana helm chart (also used by the `kube-prometheus-stack`) has the following feature:
> If the parameter `sidecar.dashboards.enabled` is set, a sidecar container is deployed in the grafana pod. This
> container watches all configmaps (or secrets) in the cluster and filters out the ones with a label as defined in
> `sidecar.dashboards.label`. The files defined in those configmaps are written to a folder and accessed by grafana.
> Changes to the configmaps are monitored and the imported dashboards are deleted/updated.

The helm chart also allows to define `sidecar.dashboards.folderAnnotation`. When this is defined and the
`sidecar.dashboards.provider.foldersFromFilesStructure` is set to `true` then:
> the sidecar will look for annotation with this name to create folder and put graph here.

Using these allow to organize the Grafana dashboards into folders and show them in this way in Grafana.

For this VAP to work, you need to set the following values in grafana helm chart:
```yaml
      sidecar:
        dashboards:
          enabled: true
          label: grafana_dashboard
          labelValue: "1"
          searchNamespace: ALL
          folderAnnotation: grafana_folder
          provider:
            foldersFromFilesStructure: true
```

# Parameter used by the policy
This policy does not use parameters. The grafana folder name (the value of the `grafana_folder` annotation) must be set
to the namespace of the `ConfigMap`/`Secret`.

It can be improved to use a list of allowed folder names in a backwards compatible way (if parameter is there then use
that if not, then fall back to namespace)

# Example snippet from a valid `ConfigMap`
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: rabbitmq-overview
  namespace: rabbitmq-monitoring
  labels:
    grafana_dashboard: "1"
  annotations:
    grafana_folder: "rabbitmq-monitoring"
data:
  rabbitmq-dashboard.json: |-
    {
      "annotations": {
.
.
.
```