### Debug information

This is a useful Prometheus query to show:

* Service replicas
* Rate of invocation
* Execution time of events

http://localhost:9090/graph?g0.range_input=15m&g0.expr=gateway_service_count&g0.tab=0&g1.range_input=15m&g1.expr=rate(gateway_function_invocation_total%5B20s%5D)&g1.tab=0&g2.range_input=15m&g2.expr=gateway_functions_seconds_sum+%2F+gateway_functions_seconds_count&g2.tab=0


```
$ docker service ls -q |xargs -n 1 -I {} docker service scale {}=10;docker service scale func_gateway=1 ;
$ docker service scale func_prometheus=1 ; docker service scale func_alertmanager=1
```
