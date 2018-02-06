package RpgManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	//"bufio"
)

//In this file, we find functions that relate to meetings set up by the bots users. As of now, it contains functions that initialize a map containing the meetings by loading them from a json file, and functions
//that add, remove, sort, and display the meetings.

//Initializes manager and asks what do next
var allMeetings map[int]Meeting

const timeStamp = "Jan _2 2006, 15:04 MST"

func StartMeetings() {
	allMeetings = make(map[int]Meeting)
	initializeAllMeetings()
}

//Loads file representing meeting times and reads off the one occuring the soonest in the future. It then prompts if the user would like to see more.

//struct that contains a slice of meetings to be loaded from a json file.
type MeetingImport struct {
	Meetings []struct {
		Label    string `json:"label"`
		DateTime string `json:"date/time"`
		Comment  string `json:"comment"`
	} `json:"meetings"`
}

//struct that holds a meeting including it's time in a proper Time format
type Meeting struct {
	Label    string
	Index    int
	DateTime time.Time
	Comment  string
}

//function prints all information about every meeting
func (m Meeting) print() {
	fmt.Println(m.DateTime.Format(timeStamp))
	fmt.Println(m.Comment)
}

func (m Meeting) toString() string {
	output := "Meeting Label: " + m.Label + "\n"
	output += "Meeting occurs/occured on: "
	//output += m.DateTime.Format(timeStamp)
	output += m.DateTime.UTC().Format(timeStamp)
	output += "\n"
	output += "Comment: "
	output += m.Comment
	output += "\n"
	return output
}

//function that returns an array of time values representing each meeting
func NextMeeting() string {
	//get the time now to use in deciding which meeting is soonest
	//timeNow := time.Now()

	//search through structs
	sortMeetingsByTime()
	/*
		allMeetings[2].print()
		allMeetings[1].print()
		fmt.Println(allMeetings[2].DateTime.Format(timeStamp))
		fmt.Println(allMeetings[1].DateTime.Format(timeStamp))
		fmt.Println(allMeetings[2].DateTime.Sub(allMeetings[1].DateTime))*/
	//print one with time-now the smallest
	return searchMeetingsForSoonest().toString()
}

func initializeAllMeetings() {
	//load json file
	meetingsFile, err := os.Open("meetingTimes.json")
	if err != nil {
		log.Fatal(err)
	}
	//Read file into a big slice of bytes
	timeData, err := ioutil.ReadAll(meetingsFile)
	if err != nil {
		log.Fatal(err)
	}

	//unpack into struct after declaring said struct
	scheduled := MeetingImport{}
	if err = json.Unmarshal(timeData, &scheduled); err != nil {
		log.Fatal(err)
	}
	//Move meetings into the map
	for i, meetings := range scheduled.Meetings {
		tempt := time.Now()
		tempt, err = time.Parse(timeStamp, meetings.DateTime)
		if err != nil {
			fmt.Println(err)
		}
		temp := Meeting{meetings.Label, i, tempt, meetings.Comment} //creates a temporary Meeting using each the meetings data
		allMeetings[i] = temp                                       //assigns temporary meeting to the key in the map
	}

}

//function that searches allMeetings for the soonest one and returns it
func searchMeetingsForSoonest() Meeting {
	for _, v := range allMeetings {
		if time.Until(v.DateTime) > 0 {
			return v
			break
		}
	}
	return allMeetings[0]
}

func sortMeetingsByTime() {
	temps := make([]Meeting, len(allMeetings))
	idx := 0
	for _, v := range allMeetings {
		temps[idx] = v
		idx++
	}
	done := false
	for done == false {
		done = true
		for i := 1; i < len(temps); i++ {
			if temps[i].DateTime.Sub(temps[i-1].DateTime) < 0 {
				temp := temps[i]
				temps[i] = temps[i-1]
				temps[i-1] = temp
				done = false
				break
			}
		}
	}
	for i, meet := range temps {
		meet.Index = i
		allMeetings[i] = meet
	}
}

func ShowAllMeetings() string {
	var output string
	for k, v := range allMeetings {
		output += v.toString()
		output += "Index: "
		output += strconv.Itoa(k)
		output += "\n\n"
	}
	return output
}

//adds Meeting to the allMeetings map and to the json file
func AddMeeting(meeting string) (error, string) {
	var temp Meeting
	prevNumMeetings := len(allMeetings)
	newNumMeetings := prevNumMeetings + 1
	tokens := strings.Split(meeting, "; ") //Splits a string into the parts needed to make a meeting
	fmt.Println(len(tokens))
	tempt, err := time.Parse(timeStamp, tokens[1]) //turns the time string into a time object
	if err == nil {
		temp = Meeting{tokens[0], newNumMeetings, tempt, tokens[2]} //creates a temporary meet object according to the string pieces
		allMeetings[newNumMeetings-1] = temp
		sortMeetingsByTime()
	} else {
		fmt.Println(err)
	}
	writeMapToFile()
	return err, temp.toString()
}

func writeMapToFile() {
	Export := mapBackToStruct()
	example, err := json.MarshalIndent(Export, "", "	")
	if err != nil {
		fmt.Println(err)
	}
	example = []byte(strings.Replace(string(example), "},", "},\n", -1))
	file, err := os.OpenFile(
		"meetingTimes.json",
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Write bytes to file
	byteSlice := []byte(example)
	bytesWritten, err := file.Write(byteSlice)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote %d bytes.\n", bytesWritten)
}

func mapBackToStruct() MeetingImport {
	//First turn the map into a slice of Meetings
	temps := make([]Meeting, len(allMeetings))
	idx := 0
	for _, v := range allMeetings {
		temps[idx] = v
		idx++
	}
	//Then create a MeetingImport and put the slice into that
	Export := MeetingImport{make([]struct {
		Label    string `json:"label"`
		DateTime string `json:"date/time"`
		Comment  string `json:"comment"`
	}, len(allMeetings))}
	for i, meet := range temps {
		Export.Meetings[i].Label = meet.Label
		Export.Meetings[i].DateTime = meet.DateTime.Format(timeStamp)
		Export.Meetings[i].Comment = meet.Comment
	}
	return Export
}

func DeleteMeeting(idx int) (string, error) {
	temp := allMeetings[idx]
	if temp.Label == "" {
		return "", errors.New("Unable to find meeting with given index")
	} else {
		delete(allMeetings, idx)
	}
	if allMeetings[idx].Label != "" {
		return "", errors.New("Unable to delete meeting")
	}
	writeMapToFile()
	return temp.toString(), nil
}
