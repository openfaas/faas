"use strict"
let fs = require('fs');
let sample = require("./sample.json");
let cheerio = require('cheerio');
let Parser = require('./parser');
var request = require("request");

const getStdin = require('get-stdin');
 
getStdin().then(content => {
  let request = JSON.parse(content);
  handle(request, request.request.intent);
});

function tellWithCard(speechOutput) {
  sample.response.outputSpeech.text = speechOutput
  sample.response.card.content = speechOutput
  sample.response.card.title = "Captains";
  console.log(JSON.stringify(sample));
  process.exit(0);
}

function handle(request, intent) {
    createList((sorted) => {
        let speechOutput = "There are currently " + sorted.length + " Docker captains.";
        tellWithCard(speechOutput);
    });
}

let createList = (next) => {
  let parser = new Parser(cheerio);

  request.get("https://www.docker.com/community/docker-captains", (err, res, text) => {
      let captains = parser.parse(text);

      let valid = 0;
      let sorted = captains.sort((x,y) => {
      if(x.text > y.text) {
          return 1;
      }
      else if(x.text < y.text) {
          return -1;
      }
      return 0;
      });
      next(sorted);
  });
};
