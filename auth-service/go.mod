module auth-service

go 1.24

require (
	db-service v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.36.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3
)

require (
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
)

replace db-service => ../db-service
