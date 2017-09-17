# Using Kong as an API gateway for OpenFaaS

[Kong](https://getkong.org) is an API gateway that provides features such as security, logging, and rate limiting. By putting this in front of OpenFaaS you can quickly get access to these things and a lot more via [the many other plugins written](https://getkong.org/plugins/) for it.

Below is a demo of how you could use Kong as an authentication layer for OpenFaaS. You should be able to paste this all (from its Markdown source) into [Play With Docker](http://labs.play-with-docker.com/) to see it in action.

## Setup OpenFaaS

    docker swarm init --advertise-addr $(hostname -i)
    git clone https://github.com/alexellis/faas
    cd faas
    ./deploy_stack.sh

## Check that one of the sample functions works

    curl localhost:8080/function/func_echoit -d 'hello world'


## Setup Kong

    docker service create --network func_functions --detach=false \
        --name kong-database \
        -p 5432:5432 \
        -e "POSTGRES_USER=kong" \
        -e "POSTGRES_DB=kong" \
        postgres:9.4

    docker service create --network func_functions --detach=false \
        --restart-condition=none --name=kong-migrations \
        -e "KONG_DATABASE=postgres" \
        -e "KONG_PG_HOST=kong-database" \
        kong:latest kong migrations up

    docker service create --network func_functions --name kong \
        -e "KONG_DATABASE=postgres" \
        -e "KONG_PG_HOST=kong-database" \
        -e "KONG_PROXY_ACCESS_LOG=/dev/stdout" \
        -e "KONG_ADMIN_ACCESS_LOG=/dev/stdout" \
        -e "KONG_PROXY_ERROR_LOG=/dev/stderr" \
        -e "KONG_ADMIN_ERROR_LOG=/dev/stderr" \
        -p 8000:8000 \
        -p 8443:8443 \
        -p 8001:8001 \
        -p 8444:8444 \
        kong:latest


## Put Kong in front of a single function

    echo Waiting for Kong to be ready
    until $(curl --output /dev/null --silent --head --fail http://localhost:8001); do
        printf '.'
        sleep 2
    done

    curl -i -X POST \
      --url http://localhost:8001/apis/ \
      --data 'name=echoit' \
      --data 'uris=/echo' \
      --data 'upstream_url=http://gateway:8080/function/func_echoit'

    curl localhost:8000/echo -d 'hello there'

## or put Kong in front of all the functions

    curl -i -X POST \
      --url http://localhost:8001/apis/ \
      --data 'name=functions' \
      --data 'uris=/functs' \
      --data 'upstream_url=http://gateway:8080/function'

    curl localhost:8000/functs/func_echoit -d 'hello there'


## Add a some auth with a Kong plugin

    curl -i -X POST \
      --url http://localhost:8001/apis/echoit/plugins/ \
      --data 'name=key-auth'

    curl -i -X POST \
      --url http://localhost:8001/consumers/ \
      --data "username=jdoe"

    curl -i -X POST \
      --url http://localhost:8001/consumers/jdoe/key-auth/ \
      --data 'key=longsecretkey'


## Verify the plugin worked

    curl localhost:8000/echo -d 'hello there'   # no key specified

    curl localhost:8000/echo -d 'hello there' --header "apikey: badkey"

    curl localhost:8000/echo -d 'hello there' --header "apikey: longsecretkey"
