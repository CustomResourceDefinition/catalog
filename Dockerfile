FROM alpine:3

RUN apk add --no-cache -q python3 py3-yaml yq-go helm jq wget git kustomize make shellcheck check-jsonschema editorconfig-checker

ENV PATH="$PATH:/app/bin"
RUN mkdir -p /app/bin && \
    wget -qO /app/bin/convert.py https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py && \
    printf "#!/bin/sh\npython /app/helpers/convert.py \"\$1\"" > /app/bin/convert && \
    chmod +x /app/bin/convert
