#!ps1

param (
    [switch] $noAuth,
    [switch] $noHash,
    [switch] $n,
    [switch] $help,
    [switch] $h
)

if ($help -Or $h) {
    Write-Host "Usage: "
    Write-Host " [default]`tdeploy the OpenFaaS core services"
    Write-Host " -noAuth [-n]`tdisable basic authentication"
    Write-Host " -noHash`tprevents the password from being hashed (optional)"
    Write-Host " -help [-h]`tdisplays this screen"
    Exit
}

if (Get-Command docker -errorAction SilentlyContinue)
{
    docker node ls 2>&1 | out-null
    if(-Not $?)
    {
        throw "Docker not in swarm mode, please initialise the cluster (`docker swarm init`) and retry"
    }

    # AE: would be nice to avoid this dependency.
    Add-Type -AssemblyName System.Web
    $password = [System.Web.Security.Membership]::GeneratePassword(24,5)
    $secret = ""

    if (-Not $noHash)
    {
        $sha256 = [System.Security.Cryptography.HashAlgorithm]::Create('sha256')
        $hash = $sha256.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($password))

        $secret = [System.BitConverter]::ToString($hash).Replace('-', '').toLower()
    } else {
        $secret =$password 
    }

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
        $user | docker secret create $user_secret - | out-null
    }

    $password_secret = "basic-auth-password"
    docker secret inspect $password_secret 2>&1 | out-null
    if($?)
    {
        Write-Host "$password_secret secret exists"
    }
    else
    {
        $secret | docker secret create $password_secret - | out-null
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

    Write-Host "Deploying OpenFaaS core services"
    docker stack deploy func --compose-file ./docker-compose.yml  --orchestrator swarm
}
else
{
    throw "Unable to find docker command, please install Docker (https://www.docker.com/) and retry"
}

