package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	urlsign "github.com/danilomarques1/urlSign/urlSign"
)

func main() {
	lambda.Start(urlsign.SignUrl)
}
