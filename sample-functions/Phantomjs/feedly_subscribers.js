var webPage = require('webpage');
var page = webPage.create();

var feed = "http://blog.alexellis.io/rss/";
var url = "https://feedly.com/i/subscription/feed/" + feed;

page.open(url, function(status) {
  var title = page.evaluate(function(s) {
    return document.querySelector(s).innerText;
  }, ".count-followers");
  console.log(title.split(" ")[0]);
  phantom.exit();
});
