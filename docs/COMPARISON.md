
# Comparison

This page lists missing CRD validation schemas that are present in alternative catalogs. If anyone knows the source of any of these CRDs, please make a pull request with that source or create an issue explaining where or what the source is.

## [datreeio/CRDs-catalog](https://github.com/datreeio/CRDs-catalog)

### Stats

| Coverage | Schemas in theirs | Schemas in /schema | Ignored Missing Schemas |
| --- | --- | --- | --- |
| 97.29% | 2542 | 8967 | 68 |

### Missing Schemas

| addon.open-cluster-management.io | |
| --- | --- |
| addondeploymentconfig | v1alpha1 |
| addontemplate | v1alpha1 |
| clustermanagementaddon | v1alpha1 |
| managedclusteraddon | v1alpha1 |

| autoscaling.cast.ai | |
| --- | --- |
| recommendations | v1 |

| cloud.google.com | |
| --- | --- |
| computeclass | v1 |

| cluster.open-cluster-management.io | |
| --- | --- |
| addonplacementscore | v1alpha1 |
| clusterclaim | v1alpha1 |
| managedcluster | v1 |
| managedclusterset | v1beta2 |
| managedclustersetbinding | v1beta2 |
| placement | v1beta1 |
| placementdecision | v1beta1 |

| cluster.x-k8s.io | |
| --- | --- |
| cluster | v1alpha1 |
| machine | v1alpha1 |
| machineclass | v1alpha1 |
| machinedeployment | v1alpha1 |
| machineset | v1alpha1 |

| controlplane.cluster.x-k8s.io | |
| --- | --- |
| taloscontrolplane | v1alpha3 |

| crd.k8s.amazonaws.com | |
| --- | --- |
| eniconfig | v1alpha1 |

| distro.eks.amazonaws.com | |
| --- | --- |
| release | v1alpha1 |

| eks.amazonaws.com | |
| --- | --- |
| ingressclassparams | v1 |
| nodeclass | v1 |
| nodediagnostic | v1alpha1 |
| targetgroupbinding | v1 |

| gateway.networking.k8s.io | |
| --- | --- |
| backendlbpolicy | v1alpha2 |

| getambassador.io | |
| --- | --- |
| ratelimit | v1beta2 |

| grafana.integreatly.org | |
| --- | --- |
| grafana | v1alpha1 |

| helm.toolkit.fluxcd.io | |
| --- | --- |
| helmrelease | v1 |

| infrastructure.cluster.x-k8s.io | |
| --- | --- |
| devcluster | v1beta1 |
| devclustertemplate | v1beta1 |
| devmachine | v1beta1 |
| devmachinetemplate | v1beta1 |
| gcpcluster | v1alpha2 |
| gcpmachine | v1alpha2 |
| gcpmachinetemplate | v1alpha2 |

| keda.sh | |
| --- | --- |
| cloudeventsource | v1alpha1 |

| kubeflow.org | |
| --- | --- |
| poddefault | v1alpha1 |

| kyverno.io | |
| --- | --- |
| clusterephemeralreport | v1 |
| clusterpolicyreport | v1alpha2 |
| ephemeralreport | v1 |
| policyreport | v1alpha2 |

| monitoring.googleapis.com | |
| --- | --- |
| clusternodemonitoring | v1 |
| clusterpodmonitoring | v1, v1alpha1 |
| clusterrules | v1, v1alpha1 |
| globalrules | v1, v1alpha1 |
| operatorconfig | v1, v1alpha1 |
| podmonitoring | v1, v1alpha1 |
| rules | v1, v1alpha1 |

| networking.gke.io | |
| --- | --- |
| managedcertificate | v1, v1beta1, v1beta2 |
| networklogging | v1alpha1 |

| operator.open-cluster-management.io | |
| --- | --- |
| clustermanager | v1 |
| klusterlet | v1 |

| pg.percona.com | |
| --- | --- |
| perconapgcluster | v1 |
| postgrescluster | v1beta1 |

| policies.kyverno.io | |
| --- | --- |
| policyexception | v2, v2beta1 |

| talos.dev | |
| --- | --- |
| serviceaccount | v1alpha1 |

| work.open-cluster-management.io | |
| --- | --- |
| appliedmanifestwork | v1 |
| manifestwork | v1 |
| manifestworkreplicaset | v1alpha1 |

### Ignored Schemas

<details>
<summary>Click to show ignored schemas</summary>

| | | |
| --- | --- | --- |
| anywhere.eks.amazonaws.com | cluster | v1alpha3 |
| anywhere.eks.amazonaws.com | cluster | v1alpha4 |
| anywhere.eks.amazonaws.com | cluster | v1beta1 |
| anywhere.eks.amazonaws.com | clusterclass | v1alpha4 |
| anywhere.eks.amazonaws.com | clusterclass | v1beta1 |
| anywhere.eks.amazonaws.com | clusterissuer | v1 |
| anywhere.eks.amazonaws.com | clusterresourceset | v1alpha3 |
| anywhere.eks.amazonaws.com | clusterresourceset | v1alpha4 |
| anywhere.eks.amazonaws.com | clusterresourceset | v1beta1 |
| anywhere.eks.amazonaws.com | clusterresourcesetbinding | v1alpha3 |
| anywhere.eks.amazonaws.com | clusterresourcesetbinding | v1alpha4 |
| anywhere.eks.amazonaws.com | clusterresourcesetbinding | v1beta1 |
| anywhere.eks.amazonaws.com | dockercluster | v1alpha3 |
| anywhere.eks.amazonaws.com | dockercluster | v1alpha4 |
| anywhere.eks.amazonaws.com | dockercluster | v1beta1 |
| anywhere.eks.amazonaws.com | dockerclustertemplate | v1alpha4 |
| anywhere.eks.amazonaws.com | dockerclustertemplate | v1beta1 |
| anywhere.eks.amazonaws.com | dockermachine | v1alpha3 |
| anywhere.eks.amazonaws.com | dockermachine | v1alpha4 |
| anywhere.eks.amazonaws.com | dockermachine | v1beta1 |
| anywhere.eks.amazonaws.com | dockermachinepool | v1alpha3 |
| anywhere.eks.amazonaws.com | dockermachinepool | v1alpha4 |
| anywhere.eks.amazonaws.com | dockermachinepool | v1beta1 |
| anywhere.eks.amazonaws.com | dockermachinetemplate | v1alpha3 |
| anywhere.eks.amazonaws.com | dockermachinetemplate | v1alpha4 |
| anywhere.eks.amazonaws.com | dockermachinetemplate | v1beta1 |
| dex.coreos.com | authcode | v1 |
| dex.coreos.com | authrequest | v1 |
| dex.coreos.com | connector | v1 |
| dex.coreos.com | devicerequest | v1 |
| dex.coreos.com | devicetoken | v1 |
| dex.coreos.com | oauth2client | v1 |
| dex.coreos.com | offlinesessions | v1 |
| dex.coreos.com | password | v1 |
| dex.coreos.com | refreshtoken | v1 |
| dex.coreos.com | signingkey | v1 |
| infrastructure.cluster.x-k8s.io | dockercluster | v1alpha2 |
| infrastructure.cluster.x-k8s.io | dockercluster | v1alpha3 |
| infrastructure.cluster.x-k8s.io | dockercluster | v1alpha4 |
| infrastructure.cluster.x-k8s.io | dockercluster | v1beta1 |
| infrastructure.cluster.x-k8s.io | dockerclustertemplate | v1alpha4 |
| infrastructure.cluster.x-k8s.io | dockerclustertemplate | v1beta1 |
| infrastructure.cluster.x-k8s.io | dockermachine | v1alpha2 |
| infrastructure.cluster.x-k8s.io | dockermachine | v1alpha3 |
| infrastructure.cluster.x-k8s.io | dockermachine | v1alpha4 |
| infrastructure.cluster.x-k8s.io | dockermachine | v1beta1 |
| infrastructure.cluster.x-k8s.io | dockermachinepool | v1alpha3 |
| infrastructure.cluster.x-k8s.io | dockermachinepool | v1alpha4 |
| infrastructure.cluster.x-k8s.io | dockermachinepool | v1beta1 |
| infrastructure.cluster.x-k8s.io | dockermachinepooltemplate | v1beta1 |
| infrastructure.cluster.x-k8s.io | dockermachinetemplate | v1alpha2 |
| infrastructure.cluster.x-k8s.io | dockermachinetemplate | v1alpha3 |
| infrastructure.cluster.x-k8s.io | dockermachinetemplate | v1alpha4 |
| infrastructure.cluster.x-k8s.io | dockermachinetemplate | v1beta1 |
| kafka.strimzi.io | kafkatopiccontrolacls | v1alpha1 |
| kafka.strimzi.io | kafkatopiccontrolacls | v1beta1 |
| kafka.strimzi.io | strimzipodset | v1beta2 |
| mysql.presslabs.org | mysqlbackup | v2 |
| openebs.io | blockdevice | v1alpha1 |
| openebs.io | blockdeviceclaim | v1alpha1 |
| openebs.io | diskpool | v1beta1 |
| openebs.io | diskpool | v1beta2 |
| policy.linkerd.io | httproute | v1 |
| postgresql.cnpg.io | cluster | v3 |
| rbacmanager.reactiveops.io | rbacdefinitions | v1alpha1 |
| templates.kluctl.io | objecthandler | v1alpha1 |
| uds.dev | exemption | v1alpha1 |
| uds.dev | package | v1alpha1 |

</details>
