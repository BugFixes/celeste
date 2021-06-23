package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bugfixes/celeste/internal/handler"
)

func main() {
	lambda.Start(handler.Handler)
}
