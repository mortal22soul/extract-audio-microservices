module github.com/video-converter/auth

go 1.23

toolchain go1.24.2

require (
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/video-converter/shared v0.0.0
	golang.org/x/crypto v0.28.0
	google.golang.org/grpc v1.69.4
	gorm.io/driver/postgres v1.5.2
	gorm.io/gorm v1.25.4
)

replace github.com/video-converter/shared => ../../shared

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
