'use strict'
let os = require('os');
let fs = require('fs');
let util = require('util');

module.exports = (content, callback) => {
    fs.readFile("/etc/hostname", "utf8", (err, data) => {
        let val  = "";
        val += "Hostname: " + data +"\n";
        val += "Platform: " + os.platform()+"\n";
        val += "Arch: " + os.arch() + "\n";
        val += "CPU count: " + os.cpus().length+ "\n";

        val += "Uptime: " + os.uptime()+ "\n";

        if (content && content.length && content.indexOf("verbose") > -1) {
            val += util.inspect(os.cpus()) + "\n";
            val += util.inspect(os.networkInterfaces())+ "\n";
        }

        callback(null, val);
    });
};
