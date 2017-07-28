### Phantomjs function

[Phantomjs](http://phantomjs.org) is a headless web-browser used for scraping and automation testing. This function will scrape a web-page with the JavaScript sent in through the function.

Once a function has been deployed to FaaS using the UI or one of the other methods you can invoke it with a JavaScript Phantomjs file.

**Image name:** `functions/base:phantomjs`


**Example usage:**

```
$ time curl --data-binary @cnn.js http://localhost:8080/function/phantomjs

Status: success
CNN - Breaking News, Latest News and Videos

real    0m8.729s
```

This script visits the front page of CNN and once it is fully loaded - scrapes the title.

See [cnn.js](https://github.com/alexellis/faas/tree/master/sample-functions/Phantomjs/cnn.js) as an example of a Phantomjs script.

Another example script [feedly_subscribers.js](https://github.com/alexellis/faas/tree/master/sample-functions/Phantomjs/feedly_subscribers.js) gives the count of subscribers for an RSS Feed registered on Feedly.

