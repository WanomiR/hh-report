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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.Run(ctx)

	<-a.Signal()

	fmt.Println("Shutting down the app...")
	ctx, _ = context.WithTimeout(ctx, 1*time.Second)

	<-ctx.Done()
}
