"use strict"

let getStdin = require('get-stdin');

let handle = (req) => {
    console.log("Hello from a " + process.env.NODE_ENV + " machine")
};

getStdin().then(val => {
   handle(val);
}).catch(e => {
   console.error(e.stack);
});
