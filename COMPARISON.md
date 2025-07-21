
# Comparison

This page lists missing CRD validation schemas that are present in alternative catalogs. If anyone knows the source of any of these CRDs, please make a pull request with that source or create an issue explaining where or what the source is.

## [datreeio/CRDs-catalog](https://github.com/datreeio/CRDs-catalog)

| anywhere.eks.amazonaws.com | |
| --- | --- |
| cluster | v1alpha3, v1alpha4, v1beta1 |
| clusterclass | v1alpha4, v1beta1 |
| clusterissuer | v1 |
| clusterresourceset | v1alpha3, v1alpha4, v1beta1 |
| clusterresourcesetbinding | v1alpha3, v1alpha4, v1beta1 |
| dockercluster | v1alpha3, v1alpha4, v1beta1 |
| dockerclustertemplate | v1alpha4, v1beta1 |
| dockermachine | v1alpha3, v1alpha4, v1beta1 |
| dockermachinepool | v1alpha3, v1alpha4, v1beta1 |
| dockermachinetemplate | v1alpha3, v1alpha4, v1beta1 |

| cloud.google.com | |
| --- | --- |
| computeclass | v1 |

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
| backendtlspolicy | v1alpha3 |

| getambassador.io | |
| --- | --- |
| filter | v1beta2, v2, v3alpha1 |
| filterpolicy | v1beta2, v2, v3alpha1 |
| ratelimit | v1beta1, v1beta2, v2, v3alpha1 |

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
| dockerclustertemplate | v1alpha4, v1beta1 |
| dockermachinepool | v1alpha3, v1alpha4, v1beta1 |
| dockermachinepooltemplate | v1beta1 |
| gcpcluster | v1alpha2 |
| gcpmachine | v1alpha2 |
| gcpmachinetemplate | v1alpha2 |

| k8s.keycloak.org | |
| --- | --- |
| keycloak | v2alpha1 |
| keycloakrealmimport | v2alpha1 |

| kargo.akuity.io | |
| --- | --- |
| clusterpromotiontask | v1alpha1 |
| projectconfig | v1alpha1 |
| promotiontask | v1alpha1 |

| keda.sh | |
| --- | --- |
| cloudeventsource | v1alpha1 |

| kubeflow.org | |
| --- | --- |
| poddefault | v1alpha1 |

| kyverno.io | |
| --- | --- |
| admissionreport | v1alpha2, v2 |
| backgroundscanreport | v1alpha2, v2 |
| cleanuppolicy | v2, v2alpha1, v2beta1 |
| clusteradmissionreport | v1alpha2, v2 |
| clusterbackgroundscanreport | v1alpha2, v2 |
| clustercleanuppolicy | v2, v2alpha1, v2beta1 |
| clusterephemeralreport | v1 |
| clusterpolicy | v1, v2beta1 |
| clusterpolicyreport | v1alpha2 |
| clusterreportchangerequest | v1alpha2 |
| ephemeralreport | v1 |
| globalcontextentry | v2alpha1 |
| policy | v2beta1 |
| policyexception | v2, v2alpha1, v2beta1 |
| policyreport | v1alpha2 |
| reportchangerequest | v1alpha2 |
| updaterequest | v1beta1, v2 |

| monitoring.googleapis.com | |
| --- | --- |
| clusternodemonitoring | v1 |
| clusterpodmonitoring | v1, v1alpha1 |
| clusterrules | v1, v1alpha1 |
| globalrules | v1, v1alpha1 |
| operatorconfig | v1, v1alpha1 |
| podmonitoring | v1, v1alpha1 |
| rules | v1, v1alpha1 |

| mysql.presslabs.org | |
| --- | --- |
| mysqlbackup | v2 |

| networking.gke.io | |
| --- | --- |
| lbpolicy | v1 |
| managedcertificate | v1, v1beta1, v1beta2 |
| networklogging | v1alpha1 |

| networking.k8s.aws | |
| --- | --- |
| policyendpoint | v1alpha1 |

| pg.percona.com | |
| --- | --- |
| perconapgbackup | v2beta1 |
| perconapgcluster | v1, v2beta1 |
| perconapgrestore | v2beta1 |
| postgrescluster | v1beta1 |

| policy.cert-manager.io | |
| --- | --- |
| certificaterequestpolicy | v1alpha1 |

| policy.linkerd.io | |
| --- | --- |
| httproute | v1 |

| rbacmanager.reactiveops.io | |
| --- | --- |
| rbacdefinitions | v1alpha1 |

| redis.redis.opstreelabs.in | |
| --- | --- |
| redis | v1beta2 |
| rediscluster | v1beta2 |
| redisreplication | v1beta2 |
| redissentinel | v1beta2 |

| v1.edp.epam.com | |
| --- | --- |
| clusterkeycloak | v1alpha1 |
| clusterkeycloakrealm | v1alpha1 |
| keycloak | v1 |
| keycloakauthflow | v1 |
| keycloakclient | v1 |
| keycloakclientscope | v1 |
| keycloakrealm | v1 |
| keycloakrealmcomponent | v1 |
| keycloakrealmgroup | v1 |
| keycloakrealmidentityprovider | v1 |
| keycloakrealmrole | v1 |
| keycloakrealmrolebatch | v1 |
| keycloakrealmuser | v1 |
