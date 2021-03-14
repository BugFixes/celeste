package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/bugfixes/celeste/internal/celeste"
)

func main() {
	lambda.Start(celeste.Handler)
}
