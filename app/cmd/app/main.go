package main

import (
	"app/internal/app"
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	a, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	go a.Run()

	<-a.Signal()
	fmt.Println("Shutting down the app...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	<-ctx.Done()
}
