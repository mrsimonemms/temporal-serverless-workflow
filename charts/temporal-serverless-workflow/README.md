# temporal-serverless-workflow

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

Build Temporal workflows with Serverless Workflow

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Node affinity |
| autoscaling.enabled | bool | `false` | Autoscaling enabled |
| autoscaling.maxReplicas | int | `100` | Maximum replicas |
| autoscaling.minReplicas | int | `1` | Minimum replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | When to trigger a new replica |
| config.logLevel | string | `"info"` | Log level |
| fullnameOverride | string | `""` | String to fully override names |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.repository | string | `"ghcr.io/mrsimonemms/temporal-serverless-workflow"` | Image repositiory |
| image.tag | string | `""` | Image tag - defaults to the chart's `AppVersion` if not set |
| imagePullSecrets | list | `[]` | Docker registry secret names |
| livenessProbe.httpGet.path | string | `"/livez"` | Path to demonstrate app liveness |
| livenessProbe.httpGet.port | string | `"http"` | Port to demonstrate app liveness |
| nameOverride | string | `""` | String to partially override name |
| nodeSelector | object | `{}` | Node selector |
| podAnnotations | object | `{}` | Pod [annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) |
| podLabels | object | `{}` | Pod [labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) |
| podSecurityContext | object | `{}` | Pod's [security context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) |
| readinessProbe.httpGet.path | string | `"/livez"` | Path to demonstrate app readiness |
| readinessProbe.httpGet.port | string | `"http"` | Port to demonstrate app readiness |
| replicaCount | int | `1` | Number of replicas |
| resources | object | `{}` | Configure resources available |
| securityContext | object | `{}` | Container's security context |
| service.port | int | `3000` | Service's port |
| service.type | string | `"ClusterIP"` | Service's type |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.automount | bool | `true` | Automatically mount a ServiceAccount's API credentials? |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` | Node toleration |
| volumeMounts | list | `[]` | Additional volumeMounts on the output Deployment definition. |
| volumes | list | `[]` | Additional volumes on the output Deployment definition. |
| workflow.configmap | string | `nil` |  |
| workflow.inline.do[0].step1.set.userId | int | `2` |  |
| workflow.inline.do[1].wait.wait.seconds | int | `5` |  |
| workflow.inline.do[2].getUser.call | string | `"http"` |  |
| workflow.inline.do[2].getUser.with.endpoint | string | `"https://jsonplaceholder.typicode.com/users/{{ .userId }}"` |  |
| workflow.inline.do[2].getUser.with.method | string | `"get"` |  |
| workflow.inline.document.dsl | string | `"1.0.0"` |  |
| workflow.inline.document.name | string | `"example"` |  |
| workflow.inline.document.namespace | string | `"ignored"` |  |
| workflow.inline.document.summary | string | `"An example of how to use Serverless Workflow to define Temporal Workflows"` |  |
| workflow.inline.document.title | string | `"Serverless Workflow"` |  |
| workflow.inline.document.version | string | `"0.0.1"` |  |
| workflow.inline.timeout.after.minutes | int | `1` |  |

