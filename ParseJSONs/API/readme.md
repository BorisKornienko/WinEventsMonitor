build for docker alpine CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o collorpi . 
for smallest size of container add to it a binary
add ca certs from linux