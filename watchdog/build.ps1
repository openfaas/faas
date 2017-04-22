pwd
cd watchdog

docker build -t alexellis2/watchdog:windows . -f .\Dockerfile.win
docker create --name watchdog alexellis2/watchdog:windows cmd
docker cp watchdog:/go/src/github.com/alexellis/watchdog/watchdog.exe .
