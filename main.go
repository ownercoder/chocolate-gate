package main

import (
	"os"
	"log"
	"fmt"
	"io/ioutil"
	"github.com/ownercoder/chocolate-gate/authenticate"
	"gopkg.in/yaml.v2"
	"gopkg.in/telegram-bot-api.v4"
	"time"
	"errors"
	"github.com/ownercoder/chocolate-gate/asterisk"
)

const STATE_PASSWORD_REQUIRED int64 = 1
const STATE_AUTHENTICATED int64 = 2

const DEFAULT_SLEEP time.Duration = 3 * time.Second

const LOCALE_RU = "ru-RU"

type Context struct {
	ChannelID int64
	State     int64
}

var ContextList = []*Context{}
var locales map[string]map[string]string

var keyboard = tgbotapi.ReplyKeyboardMarkup{}

func main() {
	SetupLocales()

	keyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(msg(LOCALE_RU, "open_pyaterochka")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(msg(LOCALE_RU, "open_middle")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(msg(LOCALE_RU, "open_ryabinoviy")),
		),
	)

	dir, err := os.Getwd()
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot get working directory: %s", err))
	}

	authData, err := ioutil.ReadFile(fmt.Sprintf("%s/config/auth.yaml", dir))
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot read auth config"))
	}

	asteriskData, err := ioutil.ReadFile(fmt.Sprintf("%s/config/asterisk.yaml", dir))
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot read asterisk config"))
	}

	config := authenticate.Config{}
	err = yaml.Unmarshal(authData, &config)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot read config: %s", err))
	}

	asteriskConfig := asterisk.Config{}
	err = yaml.Unmarshal(asteriskData, &asteriskConfig)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot read config: %s", err))
	}

	authenticate.SetConfig(&config)
	asterisk.SetConfig(&asteriskConfig)

	StartBot(os.Getenv("TOKEN"))
}

func StartBot(token string) {
	fmt.Printf("token: %s", token)
	fmt.Println()
	if len(token) == 0 {
		log.Panic("Empty token")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(fmt.Sprintf("Cannot start telegram bot: %s", err))
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		ProcessUpdate(update, bot)
	}
}

func ProcessUpdate(update tgbotapi.Update, botApi *tgbotapi.BotAPI) {
	message := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	channelState := GetChannelState(update.Message.Chat.ID)

	if update.Message.Text == "/start" {
		message = tgbotapi.NewMessage(update.Message.Chat.ID, msg(LOCALE_RU, "welcome"))
		message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		channelState.State = STATE_PASSWORD_REQUIRED
	} else if channelState.State == STATE_PASSWORD_REQUIRED {
		success, err := TryPassword(update.Message.Text, update.Message.Chat.ID, update.Message.Chat.UserName, botApi)

		if err != nil {
			log.Printf("%s try auth with password %s", update.Message.Chat.UserName, update.Message.Text)
		}

		if success {
			channelState.State = STATE_AUTHENTICATED
			message = tgbotapi.NewMessage(update.Message.Chat.ID, msg(LOCALE_RU, "auth_success"))
			message.ReplyMarkup = keyboard
		} else {
			message = tgbotapi.NewMessage(update.Message.Chat.ID, msg(LOCALE_RU, "error_password"))
			message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}
	} else if channelState.State == STATE_AUTHENTICATED {
		success, err := OpenGate(update.Message.Text)
		if err != nil {
			log.Printf("Cannot open gate: %s", err)
		}

		if success {
			message = tgbotapi.NewMessage(update.Message.Chat.ID, msg(LOCALE_RU, "open_success"))
		} else {
			message = tgbotapi.NewMessage(update.Message.Chat.ID, msg(LOCALE_RU, "command_error"))
		}

		message.ReplyMarkup = keyboard
	}

	botApi.Send(message)
}

func TryPassword(password string, channelID int64, userName string, botApi *tgbotapi.BotAPI) (bool, error) {
	success, err := authenticate.Auth(channelID, userName, password)
	if err != nil {
		return false, err
	}

	return success, nil
}

func OpenGate(gate string) (bool, error) {
	if gate == msg(LOCALE_RU, "open_pyaterochka") {
		asterisk.Open(asterisk.GATE_PYATEROCHKA)
		time.Sleep(DEFAULT_SLEEP)
		return true, nil
	} else if gate == msg(LOCALE_RU, "open_middle") {
		asterisk.Open(asterisk.GATE_MIDDLE)
		time.Sleep(DEFAULT_SLEEP)
		return true, nil
	} else if gate == msg(LOCALE_RU, "open_ryabinoviy") {
		asterisk.Open(asterisk.GATE_RYBINOVIY)
		time.Sleep(DEFAULT_SLEEP)
		return true, nil
	}

	return false, errors.New("unknown gate")
}

func GetChannelState(channelID int64) *Context {
	for i := 0; i < len(ContextList); i++ {
		if ContextList[i].ChannelID == channelID {
			return ContextList[i]
		}
	}

	context := Context{
		ChannelID: channelID,
		State:     STATE_PASSWORD_REQUIRED,
	}

	ContextList = append(ContextList, &context)

	return &context
}

func SetupLocales() {
	locales = make(map[string]map[string]string, 1)
	ru := make(map[string]string, 8)
	ru["error_password"] = "Увы! Пароль не верный, попробуйте еще раз"
	ru["auth_success"] = "Пароль успешно прошел проверку, теперь вы можете использовать чат для открытия шлагбаума"
	ru["open_pyaterochka"] = "Открыть вьезд Пятерочка"
	ru["open_middle"] = "Открыть центральный вьезд"
	ru["open_ryabinoviy"] = "Открыть вьезд с рябинового"
	ru["command_error"] = "Не могу распознать команду"
	ru["open_success"] = "Проезжайте"
	ru["welcome"] = "Добро пожаловать в телеграм бот для управления шлагбаумом ЖК Шоколад г. Ставрополя\n" +
		"Пожалуйста введите пароль"
	locales[LOCALE_RU] = ru
}

func msg(locale, key string) string {
	if v, ok := locales[locale]; ok {
		if v2, ok := v[key]; ok {
			return v2
		}
	}
	return ""
}
