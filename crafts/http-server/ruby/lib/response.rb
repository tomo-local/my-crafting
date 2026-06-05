module HttpServer
  class Response
    def initialize(socket)
      @socket = socket
      @keep_alive = false
    end

    def set_keep_alive(value)
      @keep_alive = value
    end

    def write(status, body)
      connection = @keep_alive ? "keep-alive" : "close"
      @socket.write(
        "HTTP/1.1 #{status}\r\n" \
        "Connection: #{connection}\r\n" \
        "Content-Length: #{body.bytesize}\r\n" \
        "\r\n" \
        "#{body}"
      )
    end
  end
end
