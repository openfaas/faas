BaseFunction for Java
================

Use this FaaS function using Java.

**Build the Java function**

    $ docker build -t <your docker user name>/openfaas-java .

**Deploy the base Java function**

-   Option 1 - click *Create a new function* on the FaaS UI

-   Option 2 - use the [faas-cli](https://github.com/openfaas/faas-cli/) (experimental)

<!-- -->

    $ curl -sL cli.openfaas.com | sh

    $ faas-cli -action=deploy -image=functions/openfaas-java -name=openfaas-java
    Deployed.
    URL: http://localhost:8080/function/openfaas-java
    200 OK

**Say Hi with input**

`curl` is good to test function.

    $ curl http://localhost:8080/function/openfaas-java -d "test"

**Customize the transformation**

If you want to customise the transformation then edit the Dockerfile or the fprocess variable and create a new function.

**Remove the function**

You can remove the function with `docker service rm openfaas-java`.
