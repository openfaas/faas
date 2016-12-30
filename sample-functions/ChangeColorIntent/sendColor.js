"use strict"

var mqtt = require('mqtt');

class Send {
  constructor(topic) {
    this.topic = topic;
  }

  sendIntensity(req, done) {
    var ops = { port: 1883, host: "iot.eclipse.org" };

    var client = mqtt.connect(ops);

    client.on('connect', () => {

      let payload = req;
      let cb = () => {
        done();
      };
      client.publish(this.topic, JSON.stringify(payload), {qos: 1}, cb);
    });
  }

  sendColor(req, done) {
    var ops = { port: 1883, host: "iot.eclipse.org" };

    var client = mqtt.connect(ops);
    let cb = () => {
      done();
    };
    client.on('connect', () => {

      let payload = req;
      client.publish(this.topic, JSON.stringify(payload), {qos: 1}, cb);
    });
  }
}

module.exports = Send;