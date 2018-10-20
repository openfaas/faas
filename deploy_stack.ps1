#!ps1

if (Get-Command docker -errorAction SilentlyContinue)
{

    $user_secret = "basic-auth-user"
    docker secret inspect $user_secret 2>&1 | out-null
    if($?)
    {
        Write-Host "$user_secret secret exists"
    }
    else
    {
        $user = Read-Host 'Admin User?'
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
        $pass = Read-Host 'Password?' -AsSecureString
        [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($pass)) | docker secret create $password_secret -
    }

    Write-Host "Deploying stack"
    docker stack deploy func --compose-file ./docker-compose.yml
}
else
{
    Write-Host "Unable to find docker command, please install Docker (https://www.docker.com/) and retry"
}
