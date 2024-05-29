FROM alpine

ARG CONFIGURATION

RUN apk add python3 py3-yaml yq-go helm jq

COPY /src /app
COPY /test/src /test

RUN cat /app/*.sh > /app/main.tmp
RUN rm /app/*.sh
RUN mv /app/main.tmp /app/main.sh

RUN wget https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py -O- > /app/helpers/convert.py 2>/dev/null

CMD tail -f /dev/null
