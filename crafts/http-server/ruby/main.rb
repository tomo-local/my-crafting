require 'socket'
require 'logger'

def handle_conn(socket, logger)
  body = "Hello, World!"
  response = <<~HTTP
  HTTP/1.1 200 OK
  Content-Length: #{body.length}

  #{body}
  HTTP

  socket.write(response)
rescue => e
  logger.error("Handle error", exception: e)
ensure
  socket.close
end


logger = Logger.new($stdout)
server = TCPServer.open(8080)
logger.info("Server listening on 0.0.0.0:8080")

begin
  loop do
    socket = server.accept
    logger.info("Accepted connection from #{socket.remote_address.ip_address}:#{socket.remote_address.ip_port}")

    handle_conn(socket, logger)

  end
rescue Interrupt
  logger.info("Server shutting down")
ensure
  server.close
end
