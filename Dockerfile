FROM alpine:3

RUN apk add --no-cache -q python3 py3-yaml yq-go helm jq wget git kustomize make go check-jsonschema && \
    GOPATH=/usr/local/ go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

ENV PATH="$PATH:/app/bin"
RUN mkdir -p /app/bin && \
    wget -qO /app/bin/convert.py https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py && \
    printf "#!/bin/sh\npython /app/bin/convert.py \"\$1\"" > /app/bin/convert && \
    chmod +x /app/bin/convert
