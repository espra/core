# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

http: require('http')
sys: require('sys')

http.createServer (req, res) ->
  res.writeHead(200, {'Content-Type': 'text/plain'})
  res.write('Hello World')
  res.close()
.listen(8000)

sys.puts 'Server running at http://127.0.0.1:8000/'
