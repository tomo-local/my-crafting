require 'socket'
require 'logger'
require_relative 'lib/server'

LOG = Logger.new($stdout)

def handle_conn(socket)
  first_line = socket.gets.chomp

  if first_line.nil?
    raise ""
  end

  fields = first_line.split(' ')

  if fields.size != 3
    raise ""
  end


  status = "200 OK"
  body = "Hello, World!"

  case fields[1]
  when "/" then
    status = "200 OK"
    body = "Hello, World!"
  when "/about" then
    status = "200 OK"
    body = "About page"
  else
    status = "404 Not Found"
    body = "Not Found"
  end

  response = "HTTP/1.1 #{status}" + "\r\n" +
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
