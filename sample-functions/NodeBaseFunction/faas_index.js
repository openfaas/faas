"use strict"

let getStdin = require('get-stdin');

let handler = require('./handler');

getStdin().then(val => {

  let req;
  if(process.env.json) {
   req = JSON.parse(val);
  } else {
    req = val
  }

  handler(req, (err, res) => {
    if(err) {
      return console.error(err);
    }

    if(process.env.json) {
      console.log(JSON.stringify(res));
    } else {
      console.log(res);
    }
  });
}).catch(e => {
  console.error(e.stack);
});
