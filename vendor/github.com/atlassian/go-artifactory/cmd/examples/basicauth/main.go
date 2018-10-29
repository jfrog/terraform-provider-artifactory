package main

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"os"
)

func main() {
	tp := artifactory.BasicAuthTransport{
		Username: os.Getenv("ARTIFACTORY_USERNAME"),
		Password: os.Getenv("ARTIFACTORY_PASSWORD"),
	}

	client, err := artifactory.NewClient(os.Getenv("ARTIFACTORY_URL"), tp.Client())
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	_, _, err = client.System.Ping(context.Background())
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
	} else {
		fmt.Println("OK")
	}
}
