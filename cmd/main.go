package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	_ "github.com/mattn/go-sqlite3"

	"github.com/CT-IK/sobes_winter/internal/app"
	"github.com/CT-IK/sobes_winter/pkg/db"
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

	_, err = db.Initialize(db.NewProdConfig("sqlite3", "file:database.db"))
	if err != nil {
		fmt.Printf("Failed to init database: %v\n", err)
		os.Exit(-1)
	}
	defer db.Close()

	app.RegisterUserHandlers(b)
	app.RegisterAdminHandlers(b)

	fmt.Println("Bot started...")
	b.Start(ctx)
}

