const path = require('path');
const fs = require('fs');
const shell = require('shelljs');
const getStdin = require('get-stdin');


const format = `USAGE: {"code": "YOUR_CODE_GOES_HERE"}
EXAMPLE: {"code": "let test = 4; console.log(test + 2)"}`

const handle = (code) => {
  fs.writeFile(path.join(__dirname, 'sample.js'), code, (err) => {
    if(err) {
      console.error(err);
    } else {
      const results = shell.exec('node sample.js', {silent:true});
      console.log(JSON.stringify({'results': results.trim()}));
    }
  });
};

getStdin()
.then((content) => {
  let request = JSON.parse(content);
  const code = request.code;
  if (code) {
    handle(code);
  } else {
    console.error(format);
  }
})
.catch((e) => {
  console.error(e.stack);
});
