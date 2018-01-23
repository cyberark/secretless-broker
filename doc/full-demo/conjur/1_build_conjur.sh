#!/bin/bash -ex

docker-compose pull conjur

docker-compose up -d conjur
docker-compose exec conjur conjurctl wait

source ./_conjur.sh

admin_api_key=$(docker-compose exec conjur conjurctl role retrieve-key dev:user:admin | tr -d '\r')

conjur_cli "$admin_api_key" variable values add db/password "$(< ../secrets/db.password)"
conjur_cli "$admin_api_key" variable values add db/ssh_key "$(< ../secrets/id_insecure)"
conjur_cli "$admin_api_key" variable values add myapp_tls/ssl_key "$(< ../secrets/proxy_tls.key)"
