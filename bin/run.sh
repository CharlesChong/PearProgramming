go install pear/server/
go install pear/runners/
./runners -port=9090 &
./runners -master="localhost:9090" -port=9010 & 
./runners -master="localhost:9090" -port=9011 & 