cd watchdog

del watchdog.exe

docker build -t alexellis2/watchdog:windows . -f .\Dockerfile.win

docker create --name watchdog alexellis2/watchdog:windows cmd

& docker cp watchdog:/go/src/github.com/alexellis/faas/watchdog/watchdog.exe .
docker rm -f watchdog
