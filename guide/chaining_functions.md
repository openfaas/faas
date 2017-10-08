# Chaining OpenFaaS functions

We will discuss client-side piping, server-side piping and the "function director" pattern.

## Client-side piping

The easiest way to chain functions is to do it on the client-side via your application code or a `curl`.

Here is an example:

We pipe a string or file into the markdown function, then pipe it into a Slack function

```
$ curl -d "# test" localhost:8080/function/markdown | \
  curl localhost:8080/function/slack --data-binary -
```

You could also do this via code, or through the `faas-cli`:

```
$ echo "test" | faas-cli invoke markdown | \
faas-cli invoke slack
```

## Server-side access via gateway

On the server side you can access any other function by calling it on the gateway over HTTP.

### Function A calls B

Let's say we have two functions:
* geolocatecity - gives a city name for a lat/lon combo in JSON format
* findiss - finds the location of the International Space Station then pretty-prints the city name by using the `geolocatecity` function

findiss Python 2.7 handler:

```
import requests

def get_space_station_location():
    return {"lat": 0.51112, "lon": -0.1234}

def handler(st):
    location = get_space_station_location()
    r = requests.post("http://gateway:8080/function/geolocatecity", location)

    print("The ISS is over the following city: " + r.content)
```


### Function Director pattern

In the Function Director pattern - we create a "wrapper function" which can then either pipes the result of function call A into function call B or compose the results of A and B before returning a result. This approach saves on bandwidth and latency vs. client-side piping and means you can version both your connector and the functions involved.

Take our previous example:

```
$ curl -d "# test" localhost:8080/function/markdown | \
  curl localhost:8080/function/slack --data-binary -
```

markdown2slack Python 2.7 handler:

```
import requests

def handler(req):
    
    markdown = requests.post("http://gateway:8080/function/markdown", req)
    slack_result = requests.post("http://gateway:8080/function/slack", markdown.content)

    print("Slack result: " + str(slack_result.status_code))
```

Practical example:

GitHub sends a "star" event to tweetfanclub function, tweetfanclub uses get-avatar to download the user's profile picture - stores that in an S3 bucket, then invokes tweetstargazer which tweets the image. A polaroid effect is added by a "polaroid" function.

This example uses a mix of regular binaries such as ImageMagick and Python handlers generated with the FaaS-CLI.

* [GitHub to Twitter Fanclub](https://github.com/alexellis/faas-twitter-fanclub/blob/master/README.md)
