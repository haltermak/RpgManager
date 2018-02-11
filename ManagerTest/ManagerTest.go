package main

import (
	"bufio"
	"fmt"
	"github.com/haltermak/RpgManager/Manager"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	RpgManager.StartMeetings()
	RpgManager.InitDice()
	for {
		input, _ := reader.ReadString('\n')
		input = strings.Replace(input, "\n", "", -1)
		if strings.HasPrefix(input, "/shutDown") {
			os.Exit(0)
		}

		if strings.HasPrefix(input, "/roll ") || strings.HasPrefix(input, "/r ") {
			rollString := input
			rollString = strings.TrimPrefix(rollString, "/roll ")
			rollString = strings.TrimPrefix(rollString, "/r ")

			rollstring, err := RpgManager.RollDice(rollString)
			if err != nil {
				fmt.Println("I can't roll that")
				fmt.Println(err)
			} else {
				fmt.Println(rollstring)
			}
		}
		if strings.HasPrefix(input, "/addMeeting ") {
			meetingString := strings.TrimPrefix(input, "/addMeeting ")
			err, temp := RpgManager.AddMeeting(meetingString)
			if err != nil {
				fmt.Println("Error processing meeting")
			} else {
				fmt.Println(temp)
			}
		}
		if strings.HasPrefix(input, "/removeMeeting ") {
			meetingString := strings.TrimPrefix(input, "/removeMeeting ")
			meetingString = strings.Replace(meetingString, " ", "", -1)
			meetingIdx, err := strconv.Atoi(meetingString)
			if err != nil {
				fmt.Println(err)
			}
			temp, err := RpgManager.DeleteMeeting(meetingIdx)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(temp)
			}
		}
		if strings.HasPrefix(input, "/allMeetings") {
			fmt.Println(input)
			fmt.Println(RpgManager.ShowAllMeetings(input))
		}
		if strings.HasPrefix(input, "/nextMeeting") {
			fmt.Println(input)
			fmt.Println(RpgManager.NextMeeting(input))
		}
	}
}
