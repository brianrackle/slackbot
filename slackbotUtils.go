package slackbot

import (
	"errors"
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

func GetRegexNamedCaptures(api *SlackAPI, task *RegxTask, user *slack.User, text string) (*slack.User, error) {
	submatches := task.Regx.FindStringSubmatch(text)
	if len(submatches) <= 1 {
		return nil, errors.New("Unable to find any submatches in text")
	}

	namesBarkAt := task.Regx.SubexpNames()
	captures := mapNamedCaptures(submatches, namesBarkAt)

	targetUser, _ := api.Client.GetUserInfo(captures["ID"])
	if targetUser.IsBot {
		return targetUser, errors.New("Target user is a bot")
	}

	return targetUser, nil
}

//MessageRegexResponseTask a function for performing an action based on a regex match
func MessageRegexResponseTask(api *SlackAPI, task *RegxTask, user *slack.User, text string) bool {
	targetUser, err := GetRegexNamedCaptures(api, task, user, text)
	if err != nil {
		return false
	}

	parameters := slack.NewPostMessageParameters()
	parameters.AsUser = true
	api.Client.PostMessage(targetUser.Name, task.TaskMessage, parameters)
	api.Client.PostMessage(user.ID, fmt.Sprintf(task.ResponseMessage, targetUser.Name), parameters)
	return true
}

func mapNamedCaptures(matches, names []string) map[string]string {
	matches, names = matches[1:], names[1:]
	result := make(map[string]string, len(matches))
	for i, name := range names {
		result[name] = matches[i]
	}
	return result
}
