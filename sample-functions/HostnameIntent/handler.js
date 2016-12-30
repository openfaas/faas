"use strict"
let fs = require('fs');
let sample = require("./sample.json");

var content = '';
process.stdin.resume();
process.stdin.on('data', function(buf) { content += buf.toString(); });
process.stdin.on('end', function() {
    fs.readFile("/etc/hostname", "utf8", (err, data) => {
      if(err) {
        return console.log(err);
      }
//      console.log(content);

      sample.response.outputSpeech.text = "Your hostname is: " + data;
      sample.response.card.content = "Your hostname is: "+ data
      sample.response.card.title = "Your hostname";
      console.log(JSON.stringify(sample));
      process.exit(0);
   });
});

