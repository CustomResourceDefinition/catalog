
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

| app.redislabs.com | |
| --- | --- |
| redisenterpriseactiveactivedatabase | v1alpha1 |
| redisenterprisecluster | v1, v1alpha1 |
| redisenterprisedatabase | v1alpha1 |
| redisenterpriseremotecluster | v1alpha1 |

| bootstrap.cluster.x-k8s.io | |
| --- | --- |
| talosconfig | v1alpha2, v1alpha3 |
| talosconfigtemplate | v1alpha2, v1alpha3 |

| cert-manager.k8s.cloudflare.com | |
| --- | --- |
| originissuer | v1 |

| cloud.google.com | |
| --- | --- |
| backendconfig | v1, v1beta1 |
| computeclass | v1 |

| cloudsql.cloud.google.com | |
| --- | --- |
| authproxyworkload | v1 |

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

| core.cnrm.cloud.google.com | |
| --- | --- |
| configconnector | v1beta1 |
| configconnectorcontext | v1beta1 |

| crd.k8s.amazonaws.com | |
| --- | --- |
| eniconfig | v1alpha1 |

| customize.core.cnrm.cloud.google.com | |
| --- | --- |
| controllerresource | v1alpha1, v1beta1 |
| mutatingwebhookconfigurationcustomization | v1alpha1, v1beta1 |
| namespacedcontrollerresource | v1alpha1, v1beta1 |
| validatingwebhookconfigurationcustomization | v1alpha1, v1beta1 |

| distro.eks.amazonaws.com | |
| --- | --- |
| release | v1alpha1 |

| eks.amazonaws.com | |
| --- | --- |
| ingressclassparams | v1 |
| nodeclass | v1 |
| nodediagnostic | v1alpha1 |
| targetgroupbinding | v1 |

| fluentbit.fluent.io | |
| --- | --- |
| clusterfilter | v1alpha2 |
| clusterfluentbitconfig | v1alpha2 |
| clusterinput | v1alpha2 |
| clustermultilineparser | v1alpha2 |
| clusteroutput | v1alpha2 |
| clusterparser | v1alpha2 |
| collector | v1alpha2 |
| filter | v1alpha2 |
| fluentbit | v1alpha2 |
| fluentbitconfig | v1alpha2 |
| multilineparser | v1alpha2 |
| output | v1alpha2 |
| parser | v1alpha2 |

| fluentd.fluent.io | |
| --- | --- |
| clusterfilter | v1alpha1 |
| clusterfluentdconfig | v1alpha1 |
| clusterinput | v1alpha1 |
| clusteroutput | v1alpha1 |
| filter | v1alpha1 |
| fluentd | v1alpha1 |
| fluentdconfig | v1alpha1 |
| input | v1alpha1 |
| output | v1alpha1 |

| fluxcd.controlplane.io | |
| --- | --- |
| fluxinstance | v1 |
| fluxreport | v1 |
| resourceset | v1 |
| resourcesetinputprovider | v1 |

| gateway.envoyproxy.io | |
| --- | --- |
| backend | v1alpha1 |
| backendtrafficpolicy | v1alpha1 |
| clienttrafficpolicy | v1alpha1 |
| envoyextensionpolicy | v1alpha1 |
| envoypatchpolicy | v1alpha1 |
| envoyproxy | v1alpha1 |
| httproutefilter | v1alpha1 |
| securitypolicy | v1alpha1 |

| gateway.networking.k8s.io | |
| --- | --- |
| backendlbpolicy | v1alpha2 |
| backendtlspolicy | v1alpha3 |

| getambassador.io | |
| --- | --- |
| authservice | v3alpha1 |
| consulresolver | v3alpha1 |
| devportal | v1, v3alpha1 |
| filter | v1beta2, v2, v3alpha1 |
| filterpolicy | v1beta2, v2, v3alpha1 |
| host | v3alpha1 |
| kubernetesendpointresolver | v3alpha1 |
| kubernetesserviceresolver | v3alpha1 |
| listener | v3alpha1 |
| logservice | v3alpha1 |
| mapping | v3alpha1 |
| module | v3alpha1 |
| ratelimit | v1beta1, v1beta2, v2, v3alpha1 |
| ratelimitservice | v3alpha1 |
| tcpmapping | v3alpha1 |
| tlscontext | v3alpha1 |
| tracingservice | v3alpha1 |

| grafana.integreatly.org | |
| --- | --- |
| grafana | v1alpha1, v1beta1 |
| grafanaalertrulegroup | v1beta1 |
| grafanacontactpoint | v1beta1 |
| grafanadashboard | v1beta1 |
| grafanadatasource | v1beta1 |
| grafanafolder | v1beta1 |
| grafanalibrarypanel | v1beta1 |
| grafanamutetiming | v1beta1 |
| grafananotificationpolicy | v1beta1 |
| grafananotificationpolicyroute | v1beta1 |
| grafananotificationtemplate | v1beta1 |

| groupsnapshot.storage.k8s.io | |
| --- | --- |
| volumegroupsnapshot | v1beta1 |
| volumegroupsnapshotclass | v1beta1 |
| volumegroupsnapshotcontent | v1beta1 |

| helm.toolkit.fluxcd.io | |
| --- | --- |
| helmrelease | v1 |

| image.toolkit.fluxcd.io | |
| --- | --- |
| imagepolicy | v1alpha1, v1alpha2 |
| imagerepository | v1alpha1, v1alpha2 |
| imageupdateautomation | v1alpha1, v1alpha2 |

| infrastructure.cluster.x-k8s.io | |
| --- | --- |
| bootstrapkubeconfig | v1beta1 |
| byocluster | v1beta1 |
| byoclustertemplate | v1beta1 |
| byohost | v1beta1 |
| byomachine | v1beta1 |
| byomachinetemplate | v1beta1 |
| devcluster | v1beta1 |
| devclustertemplate | v1beta1 |
| devmachine | v1beta1 |
| devmachinetemplate | v1beta1 |
| dockercluster | v1alpha2, v1alpha3, v1alpha4, v1beta1 |
| dockerclustertemplate | v1alpha4, v1beta1 |
| dockermachine | v1alpha2, v1alpha3, v1alpha4, v1beta1 |
| dockermachinepool | v1alpha3, v1alpha4, v1beta1 |
| dockermachinepooltemplate | v1beta1 |
| dockermachinetemplate | v1alpha2, v1alpha3, v1alpha4, v1beta1 |
| gcpcluster | v1alpha2 |
| gcpmachine | v1alpha2 |
| gcpmachinetemplate | v1alpha2 |
| k8sinstallerconfig | v1beta1 |
| k8sinstallerconfigtemplate | v1beta1 |
| ocicluster | v1beta1, v1beta2 |
| ociclusteridentity | v1beta1, v1beta2 |
| ociclustertemplate | v1beta1, v1beta2 |
| ocimachine | v1beta1, v1beta2 |
| ocimachinepool | v1beta1, v1beta2 |
| ocimachinepoolmachine | v1beta1, v1beta2 |
| ocimachinetemplate | v1beta1, v1beta2 |
| ocimanagedcluster | v1beta1, v1beta2 |
| ocimanagedclustertemplate | v1beta1, v1beta2 |
| ocimanagedcontrolplane | v1beta1, v1beta2 |
| ocimanagedcontrolplanetemplate | v1beta1, v1beta2 |
| ocimanagedmachinepool | v1beta1, v1beta2 |
| ocimanagedmachinepooltemplate | v1beta1, v1beta2 |
| ocivirtualmachinepool | v1beta1, v1beta2 |
| tinkerbellcluster | v1beta1 |
| tinkerbellmachine | v1beta1 |
| tinkerbellmachinetemplate | v1beta1 |

| k8s.keycloak.org | |
| --- | --- |
| keycloak | v2alpha1 |
| keycloakrealmimport | v2alpha1 |

| kargo.akuity.io | |
| --- | --- |
| clusterpromotiontask | v1alpha1 |
| freight | v1alpha1 |
| project | v1alpha1 |
| projectconfig | v1alpha1 |
| promotion | v1alpha1 |
| promotiontask | v1alpha1 |
| stage | v1alpha1 |
| warehouse | v1alpha1 |

| karpenter.sh | |
| --- | --- |
| machine | v1alpha5 |
| provisioner | v1alpha5 |

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

| mysql.presslabs.org | |
| --- | --- |
| mysqlbackup | v2 |

| network.azure.upbound.io | |
| --- | --- |
| firewallpolicyrulecollectiongroup | v1beta1 |

| networking.gke.io | |
| --- | --- |
| frontendconfig | v1beta1 |
| gcpbackendpolicy | v1 |
| gcpgatewaypolicy | v1 |
| gkenetworkparamset | v1 |
| healthcheckpolicy | v1 |
| lbpolicy | v1 |
| managedcertificate | v1, v1beta1, v1beta2 |
| network | v1 |
| networklogging | v1alpha1 |
| serviceattachment | v1, v1beta1 |
| servicenetworkendpointgroup | v1beta1 |

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

| twingate.com | |
| --- | --- |
| twingateconnector | v1beta |
| twingategroup | v1beta |
| twingateresource | v1beta |
| twingateresourceaccess | v1beta |

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
