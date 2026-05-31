require 'socket'

server = TCPServer.new(8080)
puts "started 8080"

loop do
  socket = server.accept
  puts "this is socket:"
  puts socket
  puts "================"

  request_lines =  []
  while (line = socket.gets) && line != "\r\n"
    request_lines << line
  end

  puts "this is request header"
  puts request_lines
  puts "================"

  response = <<~HTTP
    HTTP/1.1 200 OK
    Content-Length: 2
    Connection: close

    OK

  HTTP

  socket.write(response)
  socket.close

end
