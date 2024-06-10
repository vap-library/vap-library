# Description
This policy enforces specific fields for flux [HelmRelease](https://fluxcd.io/flux/components/helm/helmreleases/)
resources based on defined values in a parameter.

When there is no parameter custom resource the policy denys.

# Parameter used by the policy
The policy is using a mandatory custom resource (CR) kind called `VAPLibHelmReleaseFieldsParam`.

# Example parameter
```
apiVersion: vap-library.com/v1beta1
kind: VAPLibHelmReleaseFieldsParam
metadata:
  name: example
  namespace: example
spec:
  targetNamespace: app
  serviceAccountName: deployer
```
