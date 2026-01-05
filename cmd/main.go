package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/CT-IK/sobes_winter/internal/app"
	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("config/bot.env")
	if err != nil {
		fmt.Println("No bot token provided")
		os.Exit(-1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b, err := bot.New(os.Getenv("TOKEN")) 
	if err != nil {
		fmt.Printf("Failed to create bot: %v\n", err)
		os.Exit(-1)
	}

	app.RegisterUserHandlers(b)
	app.RegisterAdminHandlers(b)

	fmt.Println("Bot started...")
	b.Start(ctx)
}

