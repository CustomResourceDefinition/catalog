FROM alpine

RUN apk add python3 py3-yaml yq-go helm jq wget

COPY /src /app
RUN cat /app/*.sh > /app/main.tmp
RUN rm /app/*.sh
RUN mv /app/main.tmp /app/main.sh

COPY /test /app/test
RUN cat /app/test/prepare-*.sh > /app/test/prepare.tmp
RUN rm /app/test/prepare-*.sh
RUN mv /app/test/prepare.tmp /app/test/prepare.sh
RUN cat /app/test/verify-*.sh > /app/test/verify.tmp
RUN rm /app/test/verify-*.sh
RUN mv /app/test/verify.tmp /app/test/verify.sh

RUN wget https://raw.githubusercontent.com/yannh/kubeconform/master/scripts/openapi2jsonschema.py -O- > /app/helpers/convert.py 2>/dev/null

WORKDIR /repository

CMD nohup python3 -m http.server 2>&1 | grep -v '" 200 -' | grep -v 'Connection refused' & sleep infinity
