package main

import (
	"fmt"
	"os"
)

func main() {
	a := App{}
	a.Initialize(os.Getenv("URL_DB_USER"), os.Getenv("URL_DB_PASSWORD"), os.Getenv("URL_DB"))

	a.Run(fmt.Sprintf(":%s", os.Getenv("APP_PORT")))
}
