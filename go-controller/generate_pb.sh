#protoc -I . loadbalancing.proto --go_out=plugins=grpc:.
export PATH=$PATH::~/go/bin
protoc -I . loadbalancing.proto \
  --go_out=main --go-grpc_out=main \
  --go_opt=paths=source_relative
