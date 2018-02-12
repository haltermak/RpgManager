package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/haltermak/RpgManager/Manager"
	"github.com/haltermak/sFlags"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var (
	Token string
)

var gamesFile *os.File

const gamesFileName = "gamesDB.json"

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
	var err error
	gamesFile, err = os.Open("gamesDB.json")
	if err != nil {
		gamesFile, err = os.Create("gamesDB.json")
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile("gamesDB.json", []byte("{\n    }"), 0666)
	}
	gamesFile.Close()
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
	RpgManager.InitDice()
	initGames()

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
	writeGameMapToFile(gamesFileName, gameMap)
	gamesFile.Close()
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
		s.ChannelMessageSend(m.ChannelID, RpgManager.NextMeeting(m.Content))
	}
	if strings.HasPrefix(m.Content, "/addMeeting ") {
		meetingString := strings.TrimPrefix(m.Content, "/addMeeting ")
		err, temp := RpgManager.AddMeeting(meetingString)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error processing meeting")
		} else {
			s.ChannelMessageSend(m.ChannelID, temp)
		}
	}
	if strings.HasPrefix(m.Content, "/removeMeeting ") {
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
	if strings.HasPrefix(m.Content, "/allMeetings") {
		s.ChannelMessageSend(m.ChannelID, RpgManager.ShowAllMeetings(m.Content))
	}

	if strings.HasPrefix(m.Content, "/help") {
		s.ChannelMessageSend(m.ChannelID, "1. /roll or /r to roll dice\n2. /allMeetings to display all saved meetings. Use -tz and a time zone to specifiy a time zone\n3. /addMeeting to add a meeting using the string you provide\n/nextMeeting to display upcoming meeting. Again use -tz flag.\n/removeMeeting to delete a meeting with the given index")
	}

	if strings.HasPrefix(m.Content, "/createGame") {
		tempGame, response, err := NewGame(m)
		if err != nil {
			fmt.Println(err)
		}
		gameMap[tempGame.Id] = *tempGame
		s.ChannelMessageSend(m.ChannelID, "Game created with ID: "+response)
		writeGameMapToFile(gamesFileName, gameMap)
	}

	if strings.HasPrefix(m.Content, "/joinGame") {
		_, flags, err := sFlags.CreateFlags(m.Content)
		if err != nil {
			fmt.Println(err)
		}
		gameID, err := sFlags.FlagToInt(flags, "-gID")
		if err != nil {
			fmt.Println(err)
		}
		response, err := gameMap[gameID].addPlayer(m)
		if err != nil {
			fmt.Println(err)
		}
		s.ChannelMessageSend(m.ChannelID, response)
		writeGameMapToFile(gamesFileName, gameMap)
	}
}

func writeGameMapToFile(filename string, m map[int]Game) {
	tempbytes, err := json.MarshalIndent(m, "", "    ")
	tempbytes = []byte(strings.Replace(string(tempbytes), "},", "},\n", -1))
	byteSlice := []byte(tempbytes)
	file, err := os.OpenFile(filename, os.O_WRONLY, 0666)
	_, err = file.Write(byteSlice)
	if err != nil {
		fmt.Println(err)
	}
	file.Close()
}
