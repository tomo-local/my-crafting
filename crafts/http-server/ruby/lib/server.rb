require 'socket'
require_relative 'request'

module HttpServer
  class Server
    def initialize(port, handler)
      @addr = port
      @handler = handler
    end

    def listen_and_serve
      server = TCPServer.open(@addr)
      LOG.info("Server listening on 0.0.0.0:#{@addr}")
      loop do
        socket = server.accept
        LOG.info("Accepted connection from #{socket.remote_address.ip_address}:#{socket.remote_address.ip_port}")

        request = HttpServer::Request.parse(socket)
        LOG.info("Received #{request.inspect}")

        keep_alive = request.wants_keep_alive?
        @handler.call(socket, request)

        socket.close unless keep_alive
      end
    rescue Interrupt
      LOG.info("Server shutting down")
    ensure
      server&.close
    end
  end
end
