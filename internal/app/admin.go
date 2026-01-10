package app

import (
	"context"
	"fmt"
	"log"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/CT-IK/sobes_winter/pkg/db"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func RegisterAdminHandlers(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "yasha_forever", bot.MatchTypeCommand, adminGreetingsHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_admin_home", bot.MatchTypeExact, adminHomeButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_check_entries", bot.MatchTypeExact, checkEntriesButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_set_slots", bot.MatchTypeExact, setSlotsButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_set_slots_day", bot.MatchTypePrefix, setSlotsDayButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_set_slots_time", bot.MatchTypePrefix, setSlotButtonHandler)
}

func adminChooseContent() (string, *models.InlineKeyboardMarkup) {
	return "Выберите действие:", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Смотреть записи", CallbackData: "button_check_entries"},
				{Text: "Настроить слоты", CallbackData: "button_set_slots"},
			},
		},
	}
}

type registrationData struct {
	telegramId       string
	telegramUsername string
	datetimeBegin    time.Time
	datetimeEnd      time.Time
}

func checkEntriesContent(direction string) (string, *models.InlineKeyboardMarkup) {
	rows, err := db.Get().Query("SELECT telegram_id, telegram_username, datetime_begin, datetime_end FROM registrations WHERE direction = ?", direction)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return "", nil
	}
	defer rows.Close()

	registrations := make([]registrationData, 0)
	for rows.Next() {
		registration := registrationData{}
		var datetimeBeginString string
		var datetimeEndString string

		err = rows.Scan(&registration.telegramId, &registration.telegramUsername, &datetimeBeginString, &datetimeEndString)
		if err != nil {
			log.Println("Failed to scan query: ", err)
			return "", nil
		}

		registration.datetimeBegin, _ = time.Parse("2006-01-02 15:04:05.000", datetimeBeginString)
		registration.datetimeEnd, _ = time.Parse("2006-01-02 15:04:05.000", datetimeEndString)
		registrations = append(registrations, registration)
	}

	sort.Slice(registrations, func(i, j int) bool {
		return registrations[i].datetimeBegin.Before(registrations[j].datetimeBegin)
	})

	lastDate := time.Time{}
	var outputMessage strings.Builder
	outputMessage.WriteString("Текущие записи:\n")

	for _, registration := range registrations {
		if currentDate := time.Date(registration.datetimeBegin.Year(), registration.datetimeBegin.Month(), registration.datetimeBegin.Day(), 15, 0, 0, 0, time.UTC); currentDate != lastDate {
			lastDate = currentDate
			fmt.Fprintf(&outputMessage, "\n%s\n", currentDate.Format("02.01.2006"))
		}

		if registration.telegramUsername != "" {
			fmt.Fprintf(&outputMessage, "с %s до %s: @%s, ID: %s\n", registration.datetimeBegin.Format("15:04"), registration.datetimeEnd.Format("15:04"), registration.telegramUsername, registration.telegramId)
		} else {
			fmt.Fprintf(&outputMessage, "с %s до %s: юзернейма нет, ID: %s\n", registration.datetimeBegin.Format("15:04"), registration.datetimeEnd.Format("15:04"), registration.telegramId)
		}
	}

	if len(registrations) == 0 {
		return "Нет записей", &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "Назад", CallbackData: "button_admin_home"}},
			},
		}
	}

	return outputMessage.String(), &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "button_admin_home"}},
		},
	}
}

func setSlotsContent(direction string) (string, *models.InlineKeyboardMarkup) {
	slots, err := db.Get().Query("SELECT datetime_begin, datetime_end FROM slots WHERE direction = ?", direction)
	if err != nil {
		log.Printf("Failed to execute SELECT query: %v\n", err)
		return "", nil
	}
	defer slots.Close()

	uniqueDatesMap := make(map[string]any)
	for slots.Next() {
		var datetimeBegin string
		var datetimeEnd string

		err = slots.Scan(&datetimeBegin, &datetimeEnd)
		if err != nil {
			log.Println("Failed to scan datetime rows")
			return "", nil
		}

		date := strings.Split(datetimeBegin, " ")[0]
		uniqueDatesMap[date] = 1
	}

	uniqueDates := slices.Collect(maps.Keys(uniqueDatesMap))
	sort.Strings(uniqueDates)

	keyboard := make([][]models.InlineKeyboardButton, 0)
	for _, date := range uniqueDates {
		parts := strings.Split(date, "-")
		readableDate := fmt.Sprintf("%s.%s.%s", parts[2], parts[1], parts[0])
		keyboard = append(keyboard, []models.InlineKeyboardButton{{Text: readableDate, CallbackData: fmt.Sprintf("button_set_slots_day_%s", date)}})
	}

	keyboard = append(keyboard, []models.InlineKeyboardButton{{Text: "Назад", CallbackData: "button_admin_home"}})
	return "Выберите дату:", &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

type slotData struct {
	taken     bool
	available bool
	timeBegin string
	timeEnd   string
}

func setSlotsDayContent(direction string, date string) (string, *models.InlineKeyboardMarkup) {
	rows, err := db.Get().Query("SELECT taken, available, datetime_begin, datetime_end FROM slots WHERE direction = ? AND datetime_begin LIKE ?", direction, date+"%")
	if err != nil {
		log.Printf("Failed to execute SELECT query: %v\n", err)
		return "", nil
	}
	defer rows.Close()

	slots := make([]slotData, 0)
	for rows.Next() {
		slot := slotData{}
		var takenString string
		var availableString string

		err = rows.Scan(&takenString, &availableString, &slot.timeBegin, &slot.timeEnd)
		if err != nil {
			log.Println("Failed to scan datetime rows")
			return "", nil
		}

		slot.taken = takenString == "1"
		slot.available = availableString == "1"
		slots = append(slots, slot)
	}

	timeKeyboard := make([][]models.InlineKeyboardButton, 0)
	for _, slot := range slots {
		t, _ := time.Parse("2006-01-02 15:04:05", slot.timeBegin)
		beginTime := t.Format("15:04")

		t, _ = time.Parse("2006-01-02 15:04:05", slot.timeEnd)
		endTime := t.Format("15:04")

		status := "🟢"
		if slot.taken {
			status = "👤"
		} else if !slot.available {
			status = "🔴"
		}

		timeKeyboard = append(timeKeyboard, []models.InlineKeyboardButton{{Text: fmt.Sprintf("%s %s - %s", status, beginTime, endTime), CallbackData: fmt.Sprintf("button_set_slots_time_%s", slot.timeBegin)}})
	}

	timeKeyboard = append(timeKeyboard, []models.InlineKeyboardButton{{Text: "Назад", CallbackData: "button_set_slots"}})
	return "Выберите время:", &models.InlineKeyboardMarkup{
		InlineKeyboard: timeKeyboard,
	}
}

func adminGreetingsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var adminCount int64
	err := db.Get().QueryRow("SELECT COUNT(*) FROM admins WHERE telegram_id = ?", update.Message.From.ID).Scan(&adminCount)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	if adminCount == 0 {
		// if user is not an admin then we just silently fall
		return
	}

	tempState := states[update.Message.From.ID]
	tempState.step = registrationStepNone
	states[update.Message.From.ID] = tempState

	b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Админский функционал",
		ChatID: update.Message.Chat.ID,
	})

	text, kb := adminChooseContent()
	sendMessage(ctx, b, update.Message.Chat.ID, text, kb)
}

func adminHomeButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	text, kb := adminChooseContent()
	editMessage(ctx, b, update.CallbackQuery.Message, text, kb)
}

func checkEntriesButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	var direction string
	err := db.Get().QueryRow("SELECT direction FROM admins WHERE telegram_id = ?", update.CallbackQuery.Message.Message.Chat.ID).Scan(&direction)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	text, kb := checkEntriesContent(direction)
	editMessage(ctx, b, update.CallbackQuery.Message, text, kb)
}

func setSlotsButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	var direction string
	err := db.Get().QueryRow("SELECT direction FROM admins WHERE telegram_id = ?", update.CallbackQuery.Message.Message.Chat.ID).Scan(&direction)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	text, kb := setSlotsContent(direction)
	editMessage(ctx, b, update.CallbackQuery.Message, text, kb)
}

func setSlotsDayButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	var direction string
	err := db.Get().QueryRow("SELECT direction FROM admins WHERE telegram_id = ?", update.CallbackQuery.Message.Message.Chat.ID).Scan(&direction)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	text, kb := setSlotsDayContent(direction, strings.TrimPrefix(update.CallbackQuery.Data, "button_set_slots_day_"))
	editMessage(ctx, b, update.CallbackQuery.Message, text, kb)
}

func setSlotButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)
	datetime := strings.TrimPrefix(update.CallbackQuery.Data, "button_set_slots_time_")

	var direction string
	err := db.Get().QueryRow("SELECT direction FROM admins WHERE telegram_id = ?", update.CallbackQuery.Message.Message.Chat.ID).Scan(&direction)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	var taken string
	var available string
	err = db.Get().QueryRow("SELECT taken, available FROM slots WHERE direction = ? AND datetime_begin = ?", direction, datetime).Scan(&taken, &available)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	if taken == "1" {
		sendMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID, "Этот слот занят, его не получится закрыть насильно, напишите занявшему с просьбой перенести собеседованиe", nil)
		return
	}

	newAvailable := "1"
	if available == "1" {
		newAvailable = "0"
	}

	_, err = db.Get().Exec("UPDATE slots SET available = ? WHERE direction = ? AND datetime_begin = ?", newAvailable, direction, datetime)
	if err != nil {
		log.Println("Failed to execute and scan UPDATE query")
		return
	}

	text, kb := setSlotsDayContent(direction, strings.Split(datetime, " ")[0])
	editMessage(ctx, b, update.CallbackQuery.Message, text, kb)
}
