"use strict"
let fs = require('fs');
let sample = require("./sample.json");
let getStdin = require("get-stdin");

getStdin().then(content => {
  let request = JSON.parse(content);
  handle(request, request.request.intent);
});

function tellWithCard(speechOutput) {
  sample.response.outputSpeech.text = speechOutput
  sample.response.card.content = speechOutput
  sample.response.card.title = "Hostname";
  console.log(JSON.stringify(sample));
  process.exit(0);
}

function handle(request, intent) {
    fs.readFile("/etc/hostname", "utf8", (err, data) => {
      if(err) {
        return console.log(err);
      }
      tellWithCard("Your hostname is " + data);
  });
};
