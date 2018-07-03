#!/bin/sh

if ! [ -x "$(command -v docker)" ]; then
  echo 'Unable to find docker command, please install Docker (https://www.docker.com/) and retry' >&2
  exit 1
fi

auth="1"

while [ ! $# -eq 0 ]
do
	case "$1" in
		--no-auth | -n)
			auth=0
			;;
    --help | -h)
			echo "Usage: \n [default]\tdeploy the OpenFaaS core services\n --no-auth [-n]\tdisable basic authentication.\n --help\tdisplays this screen"
      exit
			;;
	esac
	shift
done

# Secrets should be created even if basic-auth is disabled.
echo "Attempting to create credentials for gateway.."
echo "admin" | docker secret create basic-auth-user -
secret=$(head -c 16 /dev/random|shasum -a 256 | cut -d " " -f 1)
echo "$secret" | docker secret create basic-auth-password -
if [ $? = 0 ];
then
  echo "[Credentials]\n username: admin \n password: $secret\n echo -n "$secret" | faas-cli login --username=admin --password-stdin"
else
  echo "[Credentials]\n already exist, not creating"
fi

if [ $auth -eq "1" ];
then
  echo ""
  echo "Enabling basic authentication for gateway.."
  echo ""
else
  echo ""
  echo "Disabling basic authentication for gateway.."
  echo ""
  sed -i -r.bak 's/basic_auth: \"true\"/basic_auth: \"false\"/' docker-compose.yml
fi

echo "Deploying OpenFaaS core services"

docker stack deploy func --compose-file docker-compose.yml
