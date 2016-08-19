package slackbot

import (
	"fmt"
	"regexp"

	"github.com/nlopes/slack"
)

//RegxTask defines a bot task
type RegxTask struct {
	Regx            *regexp.Regexp
	TaskMessage     string
	ResponseMessage string
}

//MessageRegexResponseTask a function for performing an action based on a regex match
func MessageRegexResponseTask(api *SlackAPI, task *RegxTask, user *slack.User, text string) bool {
	namesBarkAt := task.Regx.SubexpNames()
	captures := mapNamedCaptures(task.Regx.FindStringSubmatch(text), namesBarkAt)
	targetUser, _ := api.client.GetUserInfo(captures["ID"])
	if !targetUser.IsBot {
		parameters := slack.NewPostMessageParameters()
		parameters.AsUser = true
		api.client.PostMessage(captures["ID"], task.TaskMessage, parameters)
		api.client.PostMessage(user.ID, fmt.Sprintf(task.ResponseMessage, targetUser.Name), parameters)
		return true
	}
	return false
}

func mapNamedCaptures(matches, names []string) map[string]string {
	matches, names = matches[1:], names[1:]
	result := make(map[string]string, len(matches))
	for i, name := range names {
		result[name] = matches[i]
	}
	return result
}
