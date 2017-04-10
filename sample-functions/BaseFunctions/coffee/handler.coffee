getStdin = require 'get-stdin'

handler = (req) -> console.log req

getStdin()
.then (val) -> handler val
.catch (e) -> console.error e.stack
