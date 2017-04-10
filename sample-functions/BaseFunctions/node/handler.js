"use strict"

let getStdin = require('get-stdin');

let handle = (req) => {
   console.log(req);
};

getStdin().then(val => {
   handle(val);
}).catch(e => {
  console.error(e.stack);
});
