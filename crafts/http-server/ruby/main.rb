require 'socket'
require 'logger'
require_relative 'lib/server'

LOG = Logger.new($stdout)

def handle_conn(socket)
  first_line = socket.gets&.chomp
  raise "connection closed" if first_line.nil?

  fields = first_line.split(' ')

  if fields.size != 3
    raise "Invalid request line: #{first_line}"
  end

  status, body = case fields[1]
  when "/" then ["200 OK", "Hello, World!"]
  when "/about" then ["200 OK", "About page"]
  else ["404 Not Found", "Not Found"]
  end

  response = "HTTP/1.1 #{status}" + "\r\n" +
    "Content-Length: #{body.bytesize}" + "\r\n" +
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
