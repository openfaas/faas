"use strict"
let fs = require('fs');
let sample = require("./sample_response.json");
let SendColor = require('./sendColor');
let sendColor = new SendColor("alexellis.io/officelights")

const getStdin = require('get-stdin');

getStdin().then(content => {
  let request = JSON.parse(content);
  handle(request, request.request.intent);
});

function tellWithCard(speechOutput, request) {
  sample.response.session = request.session;
  sample.response.outputSpeech.text = speechOutput;
  sample.response.card.content = speechOutput;
  sample.response.card.title = "Office Lights";

  console.log(JSON.stringify(sample));
  process.exit(0);
}

function handle(request, intent) {
  if(intent.name == "TurnOffIntent") {
    let req = {r:0,g:0,b:0};
    var speechOutput = "Lights off.";
    sendColor.sendColor(req, () => {
      return tellWithCard(speechOutput, request);
    });
  } else {
    let colorRequested = intent.slots.LedColor.value;
    let req = {r:0,g:0,b:0};
    if(colorRequested == "red") { 
      req.r = 255;
    } else if(colorRequested== "blue") {
        req.b = 255;
    } else if (colorRequested == "green") {
        req.g = 255;
    } else if (colorRequested == "white") {
        req.r = 255;
        req.g = 103;
        req.b = 23;
    } else {
        let msg = "I heard "+colorRequested+ " but can only show: red, green, blue and white.";
        return tellWithCard(msg, request);
    }

    sendColor.sendColor(req, () => {
      var speechOutput = "OK, " + colorRequested + ".";
      return tellWithCard(speechOutput, request);
    });
  }
}
