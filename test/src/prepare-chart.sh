echo "Setup chart ..."

cd /repository
helm package /chart
helm repo index .
