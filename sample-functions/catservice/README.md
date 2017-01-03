* Build the watchdog app which outputs fwatchdog

* Copy fwatchdog to the directory of each function you want to build.

* Create a service for each function:

```
# docker build -t catservice .
# docker service rm catservice ; docker service create --network=functions --name catservice catservice
```

* Consume it like this:

```
# curl -X POST -d @$HOME/.ssh/id_rsa.pub -H "X-Function: catservice" localhost:8080/
ssh-rsa AAAAB3NzaC1yc2....
```
