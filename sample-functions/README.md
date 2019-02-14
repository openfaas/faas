## Hello World in different languages:

For examples of hello-world in different programming languages see inside the BaseFunctions folder:

* [Base Functions](https://github.com/openfaas/faas/tree/master/sample-functions/BaseFunctions)

## Demo functions from closing keynote @ Dockercon

* Demo functions - [fass-dockercon](https://github.com/alexellis/faas-dockercon/)
* Video recording from Dockercon [on YouTube](https://youtu.be/-h2VTE9WnZs?t=15m52s)

## Sample functions from the FaaS stack

* [FaaS-And_Furious Community functions](https://github.com/faas-and-furious) (new)

> Also see the [community page](https://github.com/openfaas/faas/blob/master/community.md) for functions created by FaaS users and contributors.

Here is a list of some of the sample functions included this repository.

| Name                     | Details |
|--------------------------|-----------------------------------------                          |
| AlpineFunction           | BusyBox - a useful base image with busybox utilities pre-installed        |
| apikey-secret            | Example in Golang showing how to read a secret from a HTTP header and validate with a Swarm/Kubernetes secret |
| CaptainsIntent           | Alexa skill - find the count of Docker Captains |
| ChangeColorIntent        | Alexa skill - change the colour of IoT-connected lights |
| CHelloWorld              | Use C to build a function |
| echo                     | Uses `cat` from BusyBox to provide an echo function |
| DockerHubStats           | Golang function gives the count of repos a user has on the Docker hub |
| figlet                   | Generate ascii logos through the use of a binary |
| gif-maker                | Use gifsicle and ffmpeg packages from Alpine Linux to make gifs from video |
| HostnameIntent           | Prints the hostname of a container |
| MarkdownRender           | Use a Go function with vendoring to convert Markdown to HTML |
| Nmap                     | The network scanning tool as a binary-based function |
| NodeInfo                 | Node.js - gives CPU/network info on the current container |
| Phantomjs                | Use Phantomjs to scrape/automate web-pages |
| ResizeImageMagick        | Resizes an image using the imagemagick binary |
| SentimentAnalysis        | Perform sentiment analysis with the TextBlob library |
| WebhookStash             | Golang function provides way to capture webhooks - JSON/text/binary into the container filesystem |
| WordCountFunction        | BusyBox `wc` is exposed as a function / service through FaaS |
