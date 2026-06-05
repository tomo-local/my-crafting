require 'socket'
require 'logger'
require_relative 'lib/server'

LOG = Logger.new($stdout)

def handle_conn(socket, req)

  status, body = case req.path
  when "/" then ["200 OK", "Hello, World!"]
  when "/about" then ["200 OK", "About page"]
  else ["404 Not Found", "Not Found"]
  end

  response = "HTTP/1.1 #{status}" + "\r\n" +
    "Connection: #{req.connection}" + "\r\n" +
    "Content-Length: #{body.bytesize}" + "\r\n" +
    "\r\n" +
    "#{body}"

  socket.write(response)
rescue => e
  LOG.error("Handle error: #{e}")
end

server = HttpServer::Server.new(8080, method(:handle_conn))

server.listen_and_serve
