require 'socket'
require 'logger'
require_relative 'lib/server'

LOG = Logger.new($stdout)

def handle_conn(socket)
  body = "Hello, World!"

  response = "HTTP/1.1 200 OK" + "\r\n" +
    "Content-Length: #{body.length}" + "\r\n" +
    "\r\n" +
    "#{body}"

  socket.write(response)
rescue => e
  LOG.error("Handle error: #{e}")
ensure
  socket.close
end

server = HttpServer::Server.new(8080, method(:handle_conn))

server.listen_and_serve
