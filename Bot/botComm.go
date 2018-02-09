package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/haltermak/RpgManager/Manager"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var (
	Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	defer func() {
		if recover() != nil {
			dg.ChannelMessageSend("407899430902038541", "Bot killing itself")
		}
	}()

	//Initialize the meetings list
	RpgManager.StartMeetings()

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	//dg.ChannelMessageSend("407899430902038541", "Bot started")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	//dg.ChannelMessageSend("407899430902038541", "Bot going to sleep")
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	if strings.HasPrefix(m.Content, "/shutDown") && m.Author.Username == "CaptainJesus" {
		s.Close()
		os.Exit(0)
	}

	if strings.HasPrefix(m.Content, "/roll ") || strings.HasPrefix(m.Content, "/r ") {
		rollString := m.Content
		rollString = strings.TrimPrefix(rollString, "/roll ")
		rollString = strings.TrimPrefix(rollString, "/r ")

		rollstring, err := RpgManager.RollDice(rollString)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "I can't roll that")
			fmt.Println(err)
		} else {
			s.ChannelMessageSend(m.ChannelID, rollstring)
		}
	}
	if strings.HasPrefix(m.Content, "/nextMeeting") {
		s.ChannelMessageSend(m.ChannelID, RpgManager.NextMeeting())
	}
	if strings.HasPrefix(m.Content, "/addMeeting ") && m.Author.Username == "CaptainJesus" {
		meetingString := strings.TrimPrefix(m.Content, "/addMeeting ")
		err, temp := RpgManager.AddMeeting(meetingString)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error processing meeting")
		} else {
			s.ChannelMessageSend(m.ChannelID, temp)
		}
	}
	if strings.HasPrefix(m.Content, "/removeMeeting ") && m.Author.Username == "CaptainJesus" {
		meetingString := strings.TrimPrefix(m.Content, "/removeMeeting ")
		meetingString = strings.Replace(meetingString, " ", "", -1)
		meetingIdx, _ := strconv.Atoi(meetingString)
		temp, err := RpgManager.DeleteMeeting(meetingIdx)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error processing meeting")
		} else {
			s.ChannelMessageSend(m.ChannelID, temp)
		}
	}
	if strings.HasPrefix(m.Content, "/allMeetings") && m.Author.Username == "CaptainJesus" {
		s.ChannelMessageSend(m.ChannelID, RpgManager.ShowAllMeetings())
	}
}
