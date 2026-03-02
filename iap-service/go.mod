module github.com/meetup/iap-service

go 1.23.0

require (
	github.com/meetup/iap-api v0.0.0
	github.com/gin-gonic/gin v1.10.0
)

replace github.com/meetup/iap-api => ../iap-api
