
echo "Verifing result with known samples ..."
set -e
find /schema -type f -name "*.json" -exec sh -c 'echo diff $0 ${0/schema/verified-schema}' {} \; | sh -e
find /verified-schema -type f -name "*.json" -exec sh -c 'echo diff $0 ${0/verified-schema/schema}' {} \; | sh -e

# FIXME: handling testing differently?
# % find test/schema -type f -name "*.json" -exec sh -c 'echo diff $0 ${0/schema/verified-schema}' {} \;
# diff test/schema/traefik.io/middleware_v1alpha1.json test/verified-schema/traefik.io/middleware_v1alpha1.json
# diff test/schema/traefik.io/serverstransporttcp_v1alpha1.json test/verified-schema/traefik.io/serverstransporttcp_v1alpha1.json
# diff test/schema/traefik.io/tlsstore_v1alpha1.json test/verified-schema/traefik.io/tlsstore_v1alpha1.json
# diff test/schema/traefik.io/traefikservice_v1alpha1.json test/verified-schema/traefik.io/traefikservice_v1alpha1.json
# diff test/schema/traefik.io/tlsoption_v1alpha1.json test/verified-schema/traefik.io/tlsoption_v1alpha1.json
# diff test/schema/traefik.io/ingressroutetcp_v1alpha1.json test/verified-schema/traefik.io/ingressroutetcp_v1alpha1.json
# diff test/schema/traefik.io/ingressroute_v1alpha1.json test/verified-schema/traefik.io/ingressroute_v1alpha1.json
# diff test/schema/traefik.io/middlewaretcp_v1alpha1.json test/verified-schema/traefik.io/middlewaretcp_v1alpha1.json
# diff test/schema/traefik.io/serverstransport_v1alpha1.json test/verified-schema/traefik.io/serverstransport_v1alpha1.json
# diff test/schema/traefik.io/ingressrouteudp_v1alpha1.json test/verified-schema/traefik.io/ingressrouteudp_v1alpha1.json
# diff test/schema/traefik.containo.us/middleware_v1alpha1.json test/verified-schema/traefik.containo.us/middleware_v1alpha1.json
# diff test/schema/traefik.containo.us/tlsstore_v1alpha1.json test/verified-schema/traefik.containo.us/tlsstore_v1alpha1.json
# diff test/schema/traefik.containo.us/traefikservice_v1alpha1.json test/verified-schema/traefik.containo.us/traefikservice_v1alpha1.json
# diff test/schema/traefik.containo.us/tlsoption_v1alpha1.json test/verified-schema/traefik.containo.us/tlsoption_v1alpha1.json
# diff test/schema/traefik.containo.us/ingressroutetcp_v1alpha1.json test/verified-schema/traefik.containo.us/ingressroutetcp_v1alpha1.json
# diff test/schema/traefik.containo.us/ingressroute_v1alpha1.json test/verified-schema/traefik.containo.us/ingressroute_v1alpha1.json
# diff test/schema/traefik.containo.us/middlewaretcp_v1alpha1.json test/verified-schema/traefik.containo.us/middlewaretcp_v1alpha1.json
# diff test/schema/traefik.containo.us/serverstransport_v1alpha1.json test/verified-schema/traefik.containo.us/serverstransport_v1alpha1.json
# diff test/schema/traefik.containo.us/ingressrouteudp_v1alpha1.json test/verified-schema/traefik.containo.us/ingressrouteudp_v1alpha1.json
# diff test/schema/hub.traefik.io/accesscontrolpolicy_v1alpha1.json test/verified-schema/hub.traefik.io/accesscontrolpolicy_v1alpha1.json
# diff test/schema/hub.traefik.io/apiaccess_v1alpha1.json test/verified-schema/hub.traefik.io/apiaccess_v1alpha1.json
# diff test/schema/hub.traefik.io/edgeingress_v1alpha1.json test/verified-schema/hub.traefik.io/edgeingress_v1alpha1.json
# diff test/schema/hub.traefik.io/apiratelimit_v1alpha1.json test/verified-schema/hub.traefik.io/apiratelimit_v1alpha1.json
# diff test/schema/hub.traefik.io/apiversion_v1alpha1.json test/verified-schema/hub.traefik.io/apiversion_v1alpha1.json
# diff test/schema/hub.traefik.io/apiportal_v1alpha1.json test/verified-schema/hub.traefik.io/apiportal_v1alpha1.json
# diff test/schema/hub.traefik.io/api_v1alpha1.json test/verified-schema/hub.traefik.io/api_v1alpha1.json
