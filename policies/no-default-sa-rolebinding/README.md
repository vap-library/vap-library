# Description
This Validating Admission Policy ensures that the subjects of RoleBindings cannot include the
"default" service account. Note that this policy does not cover ClusterRoleBindings.

# Policy logic
This policy evaluates `subjects[]` in RoleBindings. For the API call to be accepted, none of the
subjects can be a ServiceAccount with the name "default".

# Parameter used by the policy
This policy does not use parameters.

# Examples
### Pass
Pass as the subject is a non-default service account.
```
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: example
  namespace: example
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: example-role
subjects:
- kind: ServiceAccount
  name: my-service-account
  namespace: example
```
### Fail
Failure as the subject is the "default" service account.
```
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: example
  namespace: example
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: example-role
subjects:
- kind: ServiceAccount
  name: default
  namespace: example
```
