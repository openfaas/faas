"use strict"


module.exports = class Parser {

  constructor(cheerio) {
    this.modules = {"cheerio": cheerio };
  }

  sanitize(handle) {
    let text = handle.toLowerCase();
    if(text[0]== "@") {
      text = text.substring(1);
    }
    if(handle.indexOf("twitter.com") > -1) {
      text = text.substring(text.lastIndexOf("\/")+1)
    }
    return {text: text, valid: text.indexOf("http") == -1};
  }

  parse(text) {
    let $ = this.modules.cheerio.load(text);
    
    let people = $("#captians .twitter_link a");
    let handles = [];
    people.each((i, person) => {
      let handle = person.attribs.href;
      handles.push(this.sanitize(handle));
    });
    return handles;
  }
};