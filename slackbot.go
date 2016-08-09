package slackbot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/nlopes/slack"
)

//RegxTask defines a bot task
type RegxTask struct {
	Regx            *regexp.Regexp
	TaskMessage     string
	ResponseMessage string
}

//Bot defines a bot
type Bot struct {
	Name           string
	Token          string
	Tasks          []RegxTask
	api            *slack.Client
	rtm            *slack.RTM
	DefaultMessage string
}

//RunBot runs the slackbot
func RunBot(bot Bot) {
	logFile, err := os.OpenFile(bot.Name+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	bot.api = slack.New(bot.Token)
	bot.rtm = bot.api.NewRTM()

	go bot.rtm.ManageConnection()
	for {
		processEvents(&bot)
	}
}

func processEvents(bot *Bot) {
	defer func() {
		if r := recover(); r != nil {
			bot.rtm.Disconnect()
			log.Println("Recovered: ", r)
		}
	}()

	for {
		event := <-bot.rtm.IncomingEvents
		switch data := event.Data.(type) {
		case *slack.MessageEvent:
			messageEvent(bot, data)
		default:
		}
	}
}

func messageEvent(bot *Bot, data *slack.MessageEvent) {
	user, err := bot.api.GetUserInfo(data.Msg.User)
	if err != nil {
		log.Println(err)
		return
	}

	msgString, _ := json.Marshal(data.Msg)
	log.Println(string(msgString))

	if len(data.Msg.BotID) == 0 && !user.IsBot {
		matchAction(bot, user, data.Msg.Text)
	}
}

func matchAction(bot *Bot, user *slack.User, text string) {
	for _, task := range bot.Tasks {
		if task.Regx.MatchString(text) {
			executeAction(bot.api, bot.rtm, &task, user, text)
			return
		}
	}
	parameters := slack.NewPostMessageParameters()
	parameters.AsUser = true
	bot.api.PostMessage(user.ID, bot.DefaultMessage, parameters)
}

func executeAction(api *slack.Client, rtm *slack.RTM, task *RegxTask, user *slack.User, text string) {
	namesBarkAt := task.Regx.SubexpNames()
	captures := mapNamedCaptures(task.Regx.FindStringSubmatch(text), namesBarkAt)
	targetUser, _ := api.GetUserInfo(captures["ID"])
	if !targetUser.IsBot {
		parameters := slack.NewPostMessageParameters()
		parameters.AsUser = true
		api.PostMessage(captures["ID"], task.TaskMessage, parameters)
		api.PostMessage(user.ID, fmt.Sprintf(task.ResponseMessage, targetUser.Name), parameters)
	}
}

func mapNamedCaptures(matches, names []string) map[string]string {
	matches, names = matches[1:], names[1:]
	result := make(map[string]string, len(matches))
	for i, name := range names {
		result[name] = matches[i]
	}
	return result
}
