## Sample functions for FaaS

Here is a list of the sample functions included this repository.

| Name                   | Details |
|------------------------|-----------------------------------------                          |
| AlpineFunction         | BusyBox - a useful base image with busybox utilities pre-installed        |
| ApiKeyProtected        | Example in Golang showing how to read X-Api-Key header |
| CaptainsIntent         | Alexa skill - find the count of Docker Captains |
| ChangeColorIntent      | Alexa skill - change the colour of IoT-connected lights |
| CatService             | Uses `cat` from BusyBox to provide an echo service |
| DockerHubStats         | Golang function gives the count of repos a user has on the Docker hub |
| HostnameIntent         | Prints the hostname of a container |
| NodeInfo               | Node.js - gives CPU/network info on the current container |
| WebhookStash           | Golang function provides way to capture webhooks - JSON/text/binary are all OK |
| WordCountFunction      | BusyBox `wc` is exposed as a function / service through FaaS |

For examples of hello-world, see inside the BaseFunctions folder:

* [Base Functions](https://github.com/alexellis/faas/tree/master/sample-functions/BaseFunctions)

