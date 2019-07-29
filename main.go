package main

import (
	"log"
	"os"
	"kubeokta/cli"

)

func main() {

	params, err := cli.Parse(os.Args)

	if err != nil {
		log.Fatal(err)
	}

	err = cli.Execute(*params)

	if err != nil {
		log.Fatal(err)
	}

}