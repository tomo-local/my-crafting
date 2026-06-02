require 'socket'
require 'logger'

def start_server(port, logger)
  TCPServer.open(port)
rescue Errno::EADDRINUSE => e
  logger.error("Failed to start server: #{e.message}")
  exit 1
end

def handle(socket, logger)
  logger.info("Accepted connection from #{socket.remote_address.ip_address}:#{socket.remote_address.ip_port}")

  body = "Hello, World!"
  response = "HTTP/1.1 200 OK\r\nContent-Length: #{body.length}\r\nConnection: close\r\n\r\n#{body}"
  socket.write(response)
rescue Errno::ECONNRESET, Errno::EPIPE, EOFError => e
  logger.warn("Client error: #{e.message}")
ensure
  socket.close
end

logger = Logger.new($stdout)
server = start_server(8080, logger)
logger.info("Server listening on 0.0.0.0:8080")

trap("INT") do
  logger.info("Server shutting down")
  server.close
  exit 0
end

loop do
  handle(server.accept, logger)
rescue Errno::EBADF
  break
end
