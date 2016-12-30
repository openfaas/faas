"use strict"
let fs = require('fs');
let sample = require("./sample.json");
let SendColor = require('./sendColor');
let sendColor = new SendColor("alexellis.io/tree1")

var content = '';
process.stdin.resume();
process.stdin.on('data', function(buf) { content += buf.toString(); });
process.stdin.on('end', function() {
  let request = JSON.parse(content);
  handle(request, request.request.intent);
});

function tellWithCard(speechOutput) {
  sample.response.outputSpeech.text = speechOutput
  sample.response.card.content = speechOutput
  sample.response.card.title = "Christmas Lights";
  console.log(JSON.stringify(sample));
  process.exit(0);
}

function handle(request,intent) {
  let colorRequested = intent.slots.LedColor.value;
  let req = {r:0,g:0,b:0};
  if(colorRequested == "red") { 
    req.r = 255;
  } else if(colorRequested== "blue") {
      req.b = 255;
  } else if (colorRequested == "green") {
      req.g = 255;
  }
  else {
      tellWithCard("I heard "+colorRequested+ " but can only do: red, green, blue.", "I heard "+colorRequested+ " but can only do: red, green, blue.");
  }
  var speechOutput = "OK, "+colorRequested+".";
  sendColor.sendColor(req, () => {
    tellWithCard(speechOutput);
  });
}