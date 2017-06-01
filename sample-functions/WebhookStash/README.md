WebhookStash
============

Example serverless function shows how to stash way contents of webhooks called via API gateway.

Each file is saved with the UNIX timestamp in nano seconds plus an extension of .txt

Example:

```
# curl -X POST -v -d @$HOME/.ssh/id_rsa.pub localhost:8080/function/func_webhookstash
```

Next, use `docker ps` to get the name of your webhookstash container.


Then if you find the replica you can check the disk:

```
# docker exec func_webhookstash.1.z054csrh70tgk9s5k4bb8uefq find
```
