package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/CT-IK/sobes_winter/pkg/db"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type registrationStep int

type registrationState struct {
	step          registrationStep
	id            int64
	username      string
	fio           string
	course        string
	direction     string
	datetimeBegin time.Time
	datetimeEnd   time.Time
}

const (
	registrationStepNone registrationStep = iota
	registrationStepFIO
	registrationStepCourse
	registrationStepDirection
	registrationStepDate
	registrationStepTime
	registrationStepConfirm
)

var states map[int64]registrationState

func RegisterUserHandlers(b *bot.Bot) {
	states = make(map[int64]registrationState)
	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand, greetingsHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_home", bot.MatchTypeExact, homeButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_my_profile", bot.MatchTypeExact, myProfileButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_register", bot.MatchTypeExact, registerButtonHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_show_registration", bot.MatchTypePrefix, showRegistrationButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_cancel_registration", bot.MatchTypePrefix, cancelRegistrationButtonHandler)

	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return update.Message != nil && states[update.Message.From.ID].step == registrationStepFIO
	}, receiveFIOHandler)

	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return update.Message != nil && states[update.Message.From.ID].step == registrationStepCourse
	}, receiveCourseHandler)

	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return update.CallbackQuery != nil && states[update.CallbackQuery.From.ID].step == registrationStepDirection
	}, receiveDirectionButtonHandler)

	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return update.CallbackQuery != nil && states[update.CallbackQuery.From.ID].step == registrationStepDate
	}, receiveDateButtonHandler)

	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return update.CallbackQuery != nil && states[update.CallbackQuery.From.ID].step == registrationStepTime
	}, receiveTimeButtonHandler)

	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return update.CallbackQuery != nil && states[update.CallbackQuery.From.ID].step == registrationStepConfirm
	}, confirmButtonHandler)
}

func chooseContent() (string, *models.InlineKeyboardMarkup) {
	return "Выберитe действие:", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Мои записи", CallbackData: "button_my_profile"},
				{Text: "Записаться", CallbackData: "button_register"},
			},
		},
	}
}

func myProfileContent(telegramId int64) (string, *models.InlineKeyboardMarkup) {
	text := "У вас нет записей"
	buttons := [][]models.InlineKeyboardButton{{{Text: "Назад", CallbackData: "button_home"}}}

	registrations, err := db.Get().Query("SELECT direction FROM registrations WHERE telegram_id = ?", telegramId)
	if err != nil {
		log.Fatal("Failed to execute SELECT query")
		return "", nil
	}

	for registrations.Next() {
		text = "Ваши записи"
		direction := ""
		registrations.Scan(&direction)
		buttons = append([][]models.InlineKeyboardButton{{{Text: direction, CallbackData: fmt.Sprintf("button_show_registration_%s_%d", direction, telegramId)}}}, buttons...)
	}

	return text, &models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}
}

func showRegistrationButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	var dateBeginStr string
	var dateEndStr string
	callbackData := strings.Split(strings.TrimPrefix(update.CallbackQuery.Data, "button_show_registration_"), "_")

	err := db.Get().QueryRow("SELECT datetime_begin, datetime_end FROM registrations WHERE direction = ? AND telegram_id = ?", callbackData[0], callbackData[1]).Scan(&dateBeginStr, &dateEndStr)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	dateBeginTime, _ := time.Parse("2006-01-02 15:04:05.000", dateBeginStr)
	dateEndTime, _ := time.Parse("2006-01-02 15:04:05.000", dateEndStr)
	date := dateBeginTime.Format("02.01.2006")
	timeBegin := dateBeginTime.Format("15:04")
	timeEnd := dateEndTime.Format("15:04")

	sendMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID, fmt.Sprintf("Направление: %s\nДата: %s\nВремя: с %s по %s", callbackData[0], date, timeBegin, timeEnd),
		&models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{{Text: "Отменить", CallbackData: fmt.Sprintf("button_cancel_registration_%s_%s", callbackData[0], callbackData[1])}}}})
}

func cancelRegistrationButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	var dateBeginStr string
	var dateEndStr string
	callbackData := strings.Split(strings.TrimPrefix(update.CallbackQuery.Data, "button_cancel_registration_"), "_")

	err := db.Get().QueryRow("SELECT datetime_begin, datetime_end FROM registrations WHERE direction = ? AND telegram_id = ?", callbackData[0], callbackData[1]).Scan(&dateBeginStr, &dateEndStr)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	dateBeginTime, _ := time.Parse("2006-01-02 15:04:05.000", dateBeginStr)
	dateEndTime, _ := time.Parse("2006-01-02 15:04:05.000", dateEndStr)
	date := dateBeginTime.Format("02.01.2006")
	timeBegin := dateBeginTime.Format("15:04")
	timeEnd := dateEndTime.Format("15:04")

	_, err = db.Get().Exec("DELETE FROM registrations WHERE direction = ? AND telegram_id = ?", callbackData[0], callbackData[1])
	if err != nil {
		log.Println("Failed to execute DELETE query")
		return
	}

	_, err = db.Get().Exec("UPDATE slots SET taken = 0 WHERE direction = ? AND datetime_begin = ? AND datetime_end = ?", callbackData[0], dateBeginStr, dateEndStr)
	if err != nil {
		log.Println("Failed to execute UPDATE query")
		return
	}

	var adminID int64
	err = db.Get().QueryRow("SELECT telegram_id FROM admins WHERE direction = ?", callbackData[0]).Scan(&adminID)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	sendMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID, "Запись отменена", nil)

	messageToAdmin := fmt.Sprintf("Запись на %s с %s до %s отменена\n", date, timeBegin, timeEnd)
	if update.CallbackQuery.From.Username == "" {
		messageToAdmin += fmt.Sprintf("ID: %d, юзернейма нет :(", update.CallbackQuery.From.ID)
	} else {
		messageToAdmin += fmt.Sprintf("ID: %d, юзернейм: @%s", update.CallbackQuery.From.ID, update.CallbackQuery.From.Username)
	}

	sendMessage(ctx, b, adminID, messageToAdmin, nil)
}

func registerContent() (string, *models.InlineKeyboardMarkup) {
	return "Введите ваше ФИО:", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "button_home"}},
		},
	}
}

func homeButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	text, kb := chooseContent()
	editMessage(ctx, b, update.CallbackQuery.Message, text, kb)
}

func myProfileButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	text, kb := myProfileContent(update.CallbackQuery.Message.Message.Chat.ID)
	editMessage(ctx, b, update.CallbackQuery.Message, text, kb)
}

func registerButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	tempState := states[update.CallbackQuery.From.ID]
	tempState.step = registrationStepFIO
	tempState.username = update.CallbackQuery.From.Username
	tempState.id = update.CallbackQuery.From.ID
	states[update.CallbackQuery.From.ID] = tempState

	text, kb := registerContent()
	editMessage(ctx, b, update.CallbackQuery.Message, text, kb)
}

func greetingsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMessage(ctx, b, update.Message.Chat.ID, "Привет! Это бот для записи на собеседования в комитеты Студенческого Совета Финансового университета.", nil)

	text, kb := chooseContent()
	sendMessage(ctx, b, update.Message.Chat.ID, text, kb)
}

func receiveFIOHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	tempState := states[update.Message.From.ID]
	tempState.step = registrationStepCourse
	tempState.fio = update.Message.Text
	states[update.Message.From.ID] = tempState

	sendMessage(ctx, b, update.Message.Chat.ID, "Ваш курс:", nil)
}

func receiveCourseHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	tempState := states[update.Message.From.ID]
	tempState.step = registrationStepDirection
	tempState.course = update.Message.Text
	states[update.Message.From.ID] = tempState

	sendMessage(ctx, b, update.Message.Chat.ID, "Направление, куда хотите податься:", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "ЦТ", CallbackData: "button_direction_ЦТ"}},
			{{Text: "Фото", CallbackData: "button_direction_Фото"}},
			{{Text: "Дизайн", CallbackData: "button_direction_Дизайн"}},
			{{Text: "F&U prod. (сценарист)", CallbackData: "button_direction_F&U prod. (сценарист)"}},
			{{Text: "F&U prod. (оператор)", CallbackData: "button_direction_F&U prod. (оператор)"}},
			{{Text: "СМИ", CallbackData: "button_direction_СМИ"}},
		},
	})
}

func receiveDirectionButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	var count int
	err := db.Get().QueryRow("SELECT COUNT(*) FROM registrations WHERE telegram_id = ? AND direction = ?", update.CallbackQuery.Message.Message.Chat.ID, strings.TrimPrefix(update.CallbackQuery.Data, "button_direction_")).Scan(&count)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	if count != 0 {
		sendMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID, "Вы уже записаны на это направление, выберите другое", nil)
		return
	}

	tempState := states[update.CallbackQuery.From.ID]
	tempState.step = registrationStepDate
	tempState.direction = strings.TrimPrefix(update.CallbackQuery.Data, "button_direction_")
	states[update.CallbackQuery.From.ID] = tempState

	slots, err := db.Get().Query("SELECT datetime_begin, datetime_end FROM slots WHERE direction = ? AND available = 1 AND taken = 0", tempState.direction)
	if err != nil {
		log.Printf("Failed to execute SELECT query: %v\n", err)
		return
	}
	defer slots.Close()

	uniqueDates := make(map[string]any)
	for slots.Next() {
		var datetimeBegin string
		var datetimeEnd string

		err = slots.Scan(&datetimeBegin, &datetimeEnd)
		if err != nil {
			log.Println("Failed to scan datetime rows")
			return
		}

		date := strings.Split(datetimeBegin, " ")[0]
		uniqueDates[date] = 1
	}

	if len(uniqueDates) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			Text:   "Нет доступных дат для записи",
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		})
		tempState.step = registrationStepDirection
		states[update.CallbackQuery.From.ID] = tempState
		return
	}

	dateKeyboard := make([][]models.InlineKeyboardButton, 0)
	for date := range uniqueDates {
		parts := strings.Split(date, "-")
		readableDate := fmt.Sprintf("%s.%s.%s", parts[2], parts[1], parts[0])
		dateKeyboard = append(dateKeyboard, []models.InlineKeyboardButton{{Text: readableDate, CallbackData: fmt.Sprintf("button_date_%s_%s", tempState.direction, date)}})
	}

	sendMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID, "Выберите дату:", &models.InlineKeyboardMarkup{InlineKeyboard: dateKeyboard})
}

func receiveDateButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	tempState := states[update.CallbackQuery.From.ID]
	tempState.step = registrationStepTime
	states[update.CallbackQuery.From.ID] = tempState

	callbackData := strings.Split(strings.TrimPrefix(update.CallbackQuery.Data, "button_date_"), "_")
	direction := callbackData[0]
	date := callbackData[1]

	slots, err := db.Get().Query("SELECT datetime_begin, datetime_end FROM slots WHERE direction = ? AND available = 1 AND taken = 0 AND datetime_begin LIKE ?", direction, date+"%")
	if err != nil {
		log.Printf("Failed to execute SELECT query: %v\n", err)
		return
	}
	defer slots.Close()

	datetimes := make(map[string]string)
	for slots.Next() {
		var datetimeBegin string
		var datetimeEnd string

		err = slots.Scan(&datetimeBegin, &datetimeEnd)
		if err != nil {
			log.Println("Failed to scan datetime rows")
			return
		}

		datetimes[datetimeBegin] = datetimeEnd
	}

	timeKeyboard := make([][]models.InlineKeyboardButton, 0)
	for begin, end := range datetimes {
		t, _ := time.Parse("2006-01-02 15:04:05.000", begin)
		beginTime := t.Format("15:04")

		t, _ = time.Parse("2006-01-02 15:04:05.000", end)
		endTime := t.Format("15:04")

		timeKeyboard = append(timeKeyboard, []models.InlineKeyboardButton{{Text: fmt.Sprintf("%s - %s", beginTime, endTime), CallbackData: fmt.Sprintf("button_time_%s_%s", begin, end)}})
	}

	sendMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID, "Выберите время:", &models.InlineKeyboardMarkup{InlineKeyboard: timeKeyboard})
}

func receiveTimeButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	tempState := states[update.CallbackQuery.From.ID]
	tempState.step = registrationStepConfirm

	callbackData := strings.Split(strings.TrimPrefix(update.CallbackQuery.Data, "button_time_"), "_")
	begin := callbackData[0]
	end := callbackData[1]

	beginTime, _ := time.Parse("2006-01-02 15:04:05.000", begin)
	endTime, _ := time.Parse("2006-01-02 15:04:05.000", end)

	tempState.datetimeBegin = beginTime
	tempState.datetimeEnd = endTime
	states[update.CallbackQuery.From.ID] = tempState

	sendMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID,
		fmt.Sprintf("Подтвердите правильность записи:\nНаправление: %s\nДата: %s\nВремя: c %s до %s\n", tempState.direction, tempState.datetimeBegin.Format("02.01.2006"), tempState.datetimeBegin.Format("15:04"), tempState.datetimeEnd.Format("15:04")),
		&models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{{Text: "Подтверждаю", CallbackData: "button_confirm"}}}})
}

func confirmButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)

	tempState := states[update.CallbackQuery.From.ID]
	tempState.step = registrationStepNone
	states[update.CallbackQuery.From.ID] = tempState

	var count int
	err := db.Get().QueryRow("SELECT COUNT(*) FROM users WHERE telegram_id = ?", tempState.id).Scan(&count)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	if count == 0 {
		_, err = db.Get().Exec("INSERT INTO users (telegram_id, telegram_username, fio, course) VALUES(?, ?, ?, ?)", tempState.id, tempState.username, tempState.fio, tempState.course)
		if err != nil {
			log.Println("Failed to execute INSERT INTO query")
			return
		}
	}

	beginTime := tempState.datetimeBegin.Format("2006-01-02 15:04:05.000")
	endTime := tempState.datetimeEnd.Format("2006-01-02 15:04:05.000")
	_, err = db.Get().Exec("INSERT INTO registrations (telegram_id, telegram_username, direction, datetime_begin, datetime_end) VALUES(?, ?, ?, ?, ?)", tempState.id, tempState.direction, beginTime, endTime)
	if err != nil {
		log.Println("Failed to execute INSERT INTO query")
		return
	}

	_, err = db.Get().Exec("UPDATE slots SET taken = 1 WHERE direction = ? AND datetime_begin = ? AND datetime_end = ?", tempState.direction, beginTime, endTime)
	if err != nil {
		log.Println("Failed to execute UPDATE query")
		return
	}

	var adminID int64
	err = db.Get().QueryRow("SELECT telegram_id FROM admins WHERE direction = ?", tempState.direction).Scan(&adminID)
	if err != nil {
		log.Println("Failed to execute and scan SELECT query")
		return
	}

	sendMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID, "Вы записаны", nil)

	messageToAdmin := fmt.Sprintf("Новая запись на %s с %s до %s\n", tempState.datetimeBegin.Format("02.01.2006"), tempState.datetimeBegin.Format("15:04"), tempState.datetimeEnd.Format("15:04"))
	if tempState.username == "" {
		messageToAdmin += fmt.Sprintf("ID: %d, юзернейма нет :(", tempState.id)
	} else {
		messageToAdmin += fmt.Sprintf("ID: %d, юзернейм: @%s", tempState.id, tempState.username)
	}

	sendMessage(ctx, b, adminID, messageToAdmin, nil)
}
