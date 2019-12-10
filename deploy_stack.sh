#!/bin/sh

if ! [ -x "$(command -v docker)" ]; then
  echo 'Unable to find docker command, please install Docker (https://www.docker.com/) and retry' >&2
  exit 1
fi

export BASIC_AUTH="true"
export AUTH_URL="http://basic-auth-plugin:8080/validate"

sha_cmd="shasum -a 256"
if ! command -v shasum >/dev/null; then
  sha_cmd="sha256sum"
fi

while [ ! $# -eq 0 ]
do
	case "$1" in
		--no-auth | -n)
			export BASIC_AUTH="false"
      export AUTH_URL=""
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
secret=$(head -c 16 /dev/urandom| $sha_cmd | cut -d " " -f 1)
echo "$secret" | docker secret create basic-auth-password -
if [ $? = 0 ];
then
  echo "[Credentials]\n username: admin \n password: $secret\n echo -n "$secret" | faas-cli login --username=admin --password-stdin"
else
  echo "[Credentials]\n already exist, not creating"
fi

if [ $BASIC_AUTH = "true" ];
then
  echo ""
  echo "Enabling basic authentication for gateway.."
  echo ""
else
  echo ""
  echo "Disabling basic authentication for gateway.."
  echo ""
fi

# Setup http signatures keys
rm signing.key > /dev/null 2>&1 || true && rm signing.key.pub > /dev/null 2>&1 || true
docker secret rm http-signing-private-key > /dev/null 2>&1 || true
docker secret rm http-signing-public-key > /dev/null 2>&1 || true

ssh-keygen -t rsa -b 2048 -N "" -m PEM -f signing.key > /dev/null 2>&1
openssl rsa -in ./signing.key -pubout -outform PEM -out signing.key.pub > /dev/null 2>&1

cat signing.key | docker secret create http-signing-private-key - > /dev/null 2>&1 || true
cat signing.key.pub | docker secret create http-signing-public-key - > /dev/null 2>&1 || true

rm signing.key || true && rm signing.key.pub || true
echo "Http encryption settings enabled..\n"

arch=$(uname -m)
case "$arch" in

"armv7l") echo "Deploying OpenFaaS core services for ARM"
          composefile="docker-compose.armhf.yml"
          ;;
"aarch64") echo "Deploying OpenFaaS core services for ARM64"
          composefile="docker-compose.arm64.yml"
          ;;
*) echo "Deploying OpenFaaS core services"
   composefile="docker-compose.yml"
   ;;
esac

docker stack deploy func --compose-file $composefile
