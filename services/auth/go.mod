module github.com/video-converter/auth

go 1.23

require (
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/redis/go-redis/v9 v9.7.3
	github.com/sirupsen/logrus v1.9.4
	github.com/video-converter/shared v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.12.1
	golang.org/x/crypto v0.28.0
	google.golang.org/grpc v1.64.0
)

replace github.com/video-converter/shared => ../../shared

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
