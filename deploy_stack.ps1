#!ps1

param (
    [switch] $noAuth,
    [switch] $n
)

if (Get-Command docker -errorAction SilentlyContinue)
{
    docker node ls 2>&1 | out-null
    if(-Not $?)
    {
        throw "Docker not in swarm mode, please initialise the cluster (`docker swarm init`) and retry"
    }

    Add-Type -AssemblyName System.Web
    $secret = [System.Web.Security.Membership]::GeneratePassword(24,5)
    $user = 'admin'

    Write-Host "Attempting to create credentials for gateway.."
    $user_secret = "basic-auth-user"
    docker secret inspect $user_secret 2>&1 | out-null
    if($?)
    {
        Write-Host "$user_secret secret exists"
    }
    else
    {
        $user | docker secret create $user_secret -
    }

    $password_secret = "basic-auth-password"
    docker secret inspect $password_secret 2>&1 | out-null
    if($?)
    {
        Write-Host "$password_secret secret exists"
    }
    else
    {
        $secret | docker secret create $password_secret -
        Write-Host "[Credentials]"
        Write-Host " username: admin"
        Write-Host " password: $secret"
        Write-Host " Write-Output `"$secret`" | faas-cli login --username=$user --password-stdin"
    }
    
    if ($noAuth -Or $n) {
        Write-Host ""
        Write-Host "Disabling basic authentication for gateway.."
        Write-Host ""
        $env:BASIC_AUTH="false";
    }
    else 
    {
        Write-Host ""
        Write-Host "Enabling basic authentication for gateway.."
        Write-Host ""
    }

    Write-Host "Deploying stack"
    docker stack deploy func --compose-file ./docker-compose.yml
}
else
{
    throw "Unable to find docker command, please install Docker (https://www.docker.com/) and retry"
}
