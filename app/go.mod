module github.com/infin8x/request-a-dasher/app

go 1.17

require (
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/infin8x/request-a-dasher/app/auth v0.0.0-20220829222615-e5d8355d8c05
)

require github.com/gorilla/mux v1.8.0

replace github.com/infin8x/request-a-dasher/app/auth => ./auth
