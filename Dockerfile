FROM alpine

ARG CONFIGURATION

RUN apk add python3 py3-yaml yq-go helm jq fdupes

COPY /src /app
COPY /${CONFIGURATION} /app/configuration.yaml
COPY /test/src /test

RUN wget https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py -O- > /app/helpers/convert.py 2>/dev/null
