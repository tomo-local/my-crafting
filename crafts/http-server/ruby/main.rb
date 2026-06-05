require 'socket'
require 'logger'
require_relative 'lib/server'

LOG = Logger.new($stdout)

def handle_conn(req, res)
  status, body = case
  when req.method == "POST" && req.content_length <= 0
    ["400 Bad Request", "Bad Request"]
  when req.method == "POST" && req.path == "/echo"
    ["200 OK", req.body]
  when req.method == "GET" && req.path == "/"
    ["200 OK", "Hello, World!"]
  when req.method == "GET" && req.path == "/about"
    ["200 OK", "About page"]
  else
    ["404 Not Found", "Not Found"]
  end

  res.write(status, body)
rescue => e
  LOG.error("Handle error: #{e}")
end

server = HttpServer::Server.new(8080, method(:handle_conn))

server.listen_and_serve
