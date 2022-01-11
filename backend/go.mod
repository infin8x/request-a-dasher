module github.com/infin8x/deliverate/backend

go 1.17

require (
	github.com/golang-jwt/jwt/v4 v4.2.0 // indirect
	github.com/infin8x/deliverate/backend/auth v0.0.0-00010101000000-000000000000
)

require github.com/gorilla/mux v1.8.0

replace github.com/infin8x/deliverate/backend/auth => ./auth
