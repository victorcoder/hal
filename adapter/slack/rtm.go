package slack

import (
	"time"

	"github.com/danryan/hal"
	"github.com/nlopes/slack"
)

func (a *adapter) startConnection() {
	chReceiver := make(chan slack.SlackEvent)
	api := slack.New(a.token)
	// api.SetDebug(true)

	users, err := api.GetUsers()
	if err != nil {
		hal.Logger.Debugf("%s\n", err)
	}

	hal.Logger.Debugf("Stored users: %s\n", a.Robot.Users.All())
	for _, user := range users {
		// retrieve the name and mention name of our bot from the server
		// if user.Id == api.Id {
		// 	a.name = user.Name
		// 	// skip adding the bot to the users map
		// 	continue
		// }
		// Initialize a newUser object in case we need it.
		newUser := hal.User{
			ID:   user.Id,
			Name: user.Name,
		}
		// Prepopulate our users map because we can easily do so.
		// If a user doesn't exist, set it.
		u, err := a.Robot.Users.Get(user.Id)
		if err != nil {
			// hal.Logger.Debugf("Stored: %s %s\n", user.Name, user.Id)
			a.Robot.Users.Set(user.Id, newUser)
		}
		// If the user doesn't match completely (say, if someone changes their name),
		// then adjust what we have stored.
		if u.Name != user.Name {
			a.Robot.Users.Set(user.Id, newUser)
		}
	}
	hal.Logger.Debugf("Stored users: %s\n", a.Robot.Users.All())

	a.wsAPI, err = api.StartRTM("", "http://"+a.team+".slack.com")
	if err != nil {
		hal.Logger.Debugf("%s\n", err)
	}

	go a.wsAPI.HandleIncomingEvents(chReceiver)
	go a.wsAPI.Keepalive(20 * time.Second)
	for {
		select {
		case msg := <-chReceiver:
			hal.Logger.Debug("Event Received: ")
			switch msg.Data.(type) {
			case slack.HelloEvent:
				// Ignore hello
			case *slack.MessageEvent:
				m := msg.Data.(*slack.MessageEvent)
				hal.Logger.Debugf("Message: %v\n", m)
				msg := a.newMessage(m)
				a.Receive(msg)
			case *slack.PresenceChangeEvent:
				m := msg.Data.(*slack.PresenceChangeEvent)
				hal.Logger.Debugf("Presence Change: %v\n", m)
			case slack.LatencyReport:
				m := msg.Data.(slack.LatencyReport)
				hal.Logger.Debugf("Current latency: %v\n", m.Value)
			default:
				hal.Logger.Debugf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}

func (a *adapter) newMessage(msg *slack.MessageEvent) *hal.Message {
	user, _ := a.Robot.Users.Get(msg.UserId)
	return &hal.Message{
		User: user,
		Room: msg.ChannelId,
		Text: msg.Text,
	}
}
