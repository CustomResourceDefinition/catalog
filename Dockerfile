FROM alpine

ARG CONFIGURATION

RUN apk add python3 py3-yaml yq-go helm jq

COPY /src /app
COPY /test /app/test

RUN cat /app/*.sh > /app/main.tmp
RUN rm /app/*.sh
RUN mv /app/main.tmp /app/main.sh

RUN wget https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py -O- > /app/helpers/convert.py 2>/dev/null

WORKDIR /repository

CMD nohup python3 -m http.server 2>&1 | grep -v '" 200 -' | grep -v 'Connection refused' & sleep infinity
