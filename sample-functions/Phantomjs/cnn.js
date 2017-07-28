var webPage = require('webpage');
var page = webPage.create();

page.open('https://www.cnn.com/', function(status) {
  console.log('Status: ' + status);
  console.log(page.title);
  // Do other things here...
  phantom.exit();

});
