'use strict'
let os = require('os');
let fs = require('fs');
const fsPromises = fs.promises
let util = require('util');

module.exports = async (event, context) => {
  let content = event.body;

  let res = await info(content)
  return context
    .status(200)
    .succeed(res)
}

async function info(content, callback) {
   let data = await fsPromises.readFile("/etc/hostname", "utf8")

   let val  = "";
   val += "Hostname: " + data +"\n";
   val += "Platform: " + os.platform()+"\n";
   val += "Arch: " + os.arch() + "\n";
   val += "CPU count: " + os.cpus().length+ "\n";

   val += "Uptime: " + os.uptime()+ "\n";

   if (content && content.length && content.indexOf("verbose") > -1) {
     val += util.inspect(os.cpus()) + "\n";
     val += util.inspect(os.networkInterfaces())+ "\n"
   }
   return val
};
