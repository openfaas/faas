# Setting up SSL with Traefik

What is Traefik?
'Træfɪk is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease' - [Traefik Homepage](http://traefik.io)

### Initial Setup
Use easily use combine Traefik and Let's Encrypt to setup SSL for your FaaS project.
The following is an example docker-compose.yml file.

```
version: "3"
services:
  # Traefik setup
  traefik:
    image: traefik:latest
    command: -c --docker=true
      --docker.swarmmode=true
      --docker.domain=traefik
      --docker.watch=true
      --web=true
      --debug=true
      --defaultEntryPoints='http,https'
      --acme=true
      --acme.domains='yourdomainhere.com,www.yourdomainhere.com'
      --acme.email=your@email.com
      --acme.ondemand=true
      --acme.onhostrule=true
      --acme.storage=/etc/traefik/acme/acme.json
      --acme.entryPoint=https
      --entryPoints='Name:https Address::443 TLS'
      --entryPoints='Name:http Address::80 Redirect.EntryPoint:https'
    ports:
      - 80:80
      - 443:443
      - 8080:8080
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
      - "acme:/etc/traefik/acme"
    networks:
      - traefik-net
    deploy:
      placement:
        constraints: [node.role == manager]

  # FaaS Gateway Setup
  gateway:
    image: functions/gateway:latest
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
    networks:
      - traefik-net
    deploy:
      labels:
        - traefik.port=8080
        - traefik.frontend.rule=Host:yourdomainhere.com,www.yourdomainhere.com
      placement:
        constraints: [node.role == manager]

  # Node Info
  nodeinfo:
    image: alexellis2/faas-nodeinfo:latest
    labels:
      function: "true"
    depends_on:
      - gateway
    networks:
      - traefik-net
    environment:
      no_proxY: "gateway"
      https_proxy: $https_proxy

networks:
  traefik-net:
    driver: overlay

volumes:
  acme:
```
#### What's happening here?
The above setup places the Traefik reverse proxy in front of the FaaS API gateway and allows you to reach it using HTTPS (port 443). Traefik
is also handling your SSL setup with Let's Encrypt! If you noticed we will be using HTTPS for all requests, as this
line `--entryPoints='Name:http Address::80 Redirect.EntryPoint:https'` allows all standard HTTP requests to be routed to port 443.

Using two labels on our gateway service, allows Traefik to find the service and route requests correctly.
`curl -X POST https://yourdomainhere.com/function/stackname_nodeinfo` will work as expected and visiting `yourdomain.com` or `www.yourdomain.com` will
bring up the FaaS Gateway frontend.
