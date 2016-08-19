package slackbot

import (
	"encoding/json"
	"log"
	"os"

	"github.com/nlopes/slack"
)

//Bot defines a bot
type Bot struct {
	Name           string
	Token          string
	MessageTasks   []func(pi *SlackAPI, data *slack.MessageEvent, user *slack.User) bool
	DefaultMessage string
}

//SlackAPI defines the slack api connections
type SlackAPI struct {
	client *slack.Client
	rtm    *slack.RTM
}

//implement heartbeat
// func healthCheck(w http.ResponseWriter, r *http.Request) {
// 	//ping the api server and if it succeeds then
// 	w.WriteHeader(http.StatusOK) //return ping time
// }

//RunBot runs the slackbot
func RunBot(bot Bot) {
	logFile, err := os.OpenFile(bot.Name+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// http.HandleFunc("/status", healthCheck)
	// http.ListenAndServe(":8080", nil)

	client := slack.New(bot.Token)
	rtm := client.NewRTM()

	api := SlackAPI{client: client, rtm: rtm}

	go api.rtm.ManageConnection()
	for {
		processEvents(&bot, &api)
	}
}

func processEvents(bot *Bot, api *SlackAPI) {
	defer func() {
		if r := recover(); r != nil {
			api.rtm.Disconnect()
			log.Println("Recovered: ", r)
		}
	}()

	for {
		event := <-api.rtm.IncomingEvents
		switch data := event.Data.(type) {
		case *slack.MessageEvent:
			messageEvent(bot, api, data)
		default:
		}
	}
}

func messageEvent(bot *Bot, api *SlackAPI, data *slack.MessageEvent) {
	user, err := api.client.GetUserInfo(data.Msg.User)
	if err != nil {
		log.Println(err)
		return
	}

	msgString, _ := json.Marshal(data.Msg)
	log.Println(string(msgString))

	if len(data.Msg.BotID) == 0 && !user.IsBot {
		executeMessageTasks(bot, api, data, user)
	}
}

func executeMessageTasks(bot *Bot, api *SlackAPI, data *slack.MessageEvent, user *slack.User) {
	success := false
	for _, task := range bot.MessageTasks {
		if task(api, data, user) {
			success = true
		}
	}

	if !success {
		parameters := slack.NewPostMessageParameters()
		parameters.AsUser = true
		api.client.PostMessage(user.ID, bot.DefaultMessage, parameters)
	}
}
