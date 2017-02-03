'use strict'
let os = require('os');
let fs = require('fs');

const getStdin = require('get-stdin');

getStdin().then((content) => {
    fs.readFile("/etc/hostname", "utf8", (err, data) => {
        console.log("Hostname: " + data);
        console.log("Platform: " + os.platform());
        console.log("Arch: " + os.arch());
        console.log("CPU count: " + os.cpus().length);
        console.log("Uptime: " + os.uptime())
        if (content && content.length && content.indexOf("verbose") > -1) {
            console.log(os.cpus());
            console.log(os.networkInterfaces());
        }
    });
});
