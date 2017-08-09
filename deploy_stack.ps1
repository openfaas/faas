#!ps1

if (Get-Command docker -errorAction SilentlyContinue)
{
    Write-Host "Deploying stack"
    docker stack deploy func --compose-file ./docker-compose.yml
}
else
{
    Write-Host "Unable to find docker command, please install Docker (https://www.docker.com/) and retry"
}
