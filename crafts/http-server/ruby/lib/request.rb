module HttpServer
  class Request
    attr_reader :method, :path, :version, :connection, :content_length, :body

    def initialize(method:, path:, version:, connection: "", content_length:, body:)
      @method = method
      @path = path
      @version = version
      @connection = connection
      @content_length = content_length
      @body = body
    end

    def self.parse(socket)
      first_line = socket.gets&.chomp
      raise "connection closed" if first_line.nil?

      fields = first_line.split(' ')
      raise "Invalid request line: #{first_line}" unless fields.size == 3

      connection = ""
      content_length = 0
      loop do
        header = socket.gets
        break if header.nil? || header == "\r\n"

        name, value = header.split(":", 2)
        next if value.nil?

        case name.downcase
        when "connection"
          connection = value.strip.downcase
        when "content-length"
          content_length = value.strip.to_i
        end
      end

      body = content_length&.positive? ? socket.read(content_length) : ""

      return new(
        method: fields[0],
        path: fields[1],
        version: fields[2],
        connection: connection,
        content_length: content_length,
        body: body
      )
    end

    def wants_keep_alive?
      case @version
      when "HTTP/1.1"
        return @connection != "close"
      when "HTTP/1.0"
        return @connection == "keep-alive"
      else
        return false
      end
    end

    def inspect
      return "#<Request method=#{@method} path=#{@path} version=#{@version}>"
    end

  end
end
