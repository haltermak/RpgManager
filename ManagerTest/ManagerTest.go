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
	for {
		input, _ := reader.ReadString('\n')
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
		if strings.HasPrefix(input, "/nextMeeting") {
			fmt.Println(RpgManager.NextMeeting())
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
			meetingIdx, _ := strconv.Atoi(meetingString)
			temp, err := RpgManager.DeleteMeeting(meetingIdx)
			if err != nil {
				fmt.Println("Error processing meeting")
			} else {
				fmt.Println(temp)
			}
		}
		if strings.HasPrefix(input, "/allMeetings") {
			fmt.Println(RpgManager.ShowAllMeetings())
		}
	}
}
