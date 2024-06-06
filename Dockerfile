FROM alpine

RUN apk add python3 py3-yaml yq-go helm jq wget git kustomize make shellcheck

CMD sleep infinity
