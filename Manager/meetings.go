package RpgManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/haltermak/sFlags"
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
var timeZones map[string]*time.Location

const timeStamp = "Jan _2 2006, 15:04 MST"

func StartMeetings() {
	allMeetings = make(map[int]Meeting)
	timeZones = make(map[string]*time.Location)
	fmt.Println("Building?")
	var err error
	timeZones["PST"], err = time.LoadLocation("America/Los_Angeles")
	timeZones["PDT"], err = time.LoadLocation("America/Los_Angeles")
	timeZones["CET"], err = time.LoadLocation("Europe/Berlin")
	timeZones["MST"], err = time.LoadLocation("America/Phoenix")
	timeZones["MDT"], err = time.LoadLocation("America/Phoenix")
	timeZones["CEST"], err = time.LoadLocation("Europe/Berlin")
	if err != nil {
		fmt.Println(err)
	}
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

func (m Meeting) toStringInLoc(loc string) string {
	output := "Meeting Label: " + m.Label + "\n"
	output += "Meeting occurs/occured on: "
	//output += m.DateTime.Format(timeStamp)
	if loc != "" {
		if timeZones[loc] != nil {
			output += m.DateTime.In(timeZones[loc]).Format(timeStamp)
		} else {
			output += m.DateTime.UTC().Format(timeStamp)
		}
	} else {
		output += m.DateTime.UTC().Format(timeStamp)
	}
	output += "\n"
	output += "Comment: "
	output += m.Comment
	output += "\n"
	return output
}

//function that returns an array of time values representing each meeting
func NextMeeting(command string) string {
	initializeAllMeetings()
	_, flags, err := sFlags.CreateFlags(command)
	if err != nil {
		fmt.Println(err)
	}
	//get the time now to use in deciding which meeting is soonest
	//timeNow := time.Now()

	//search through structs
	sortMeetingsByTime()
	return searchMeetingsForSoonest().toStringInLoc(flags["-tz"])

}

func initializeAllMeetings() {
	//load json file
	meetingsFile, err := os.Open("meetingTimes.json")
	if err != nil {
		meetingsFile, err = os.Create("meetingTimes.json")
		if err != nil {
			log.Fatal(err)
		}
		AddMeeting("Dummy Session; Jan 1 1970, 00:00 UTC; Testing")
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
		tZone := meetings.DateTime
		tZoneS := strings.Split(tZone, " ")
		tZone = tZoneS[len(tZoneS)-1]
		tempt, err = time.ParseInLocation(timeStamp, meetings.DateTime, timeZones[tZone])
		if err != nil {
			fmt.Println(err)
		}
		temp := Meeting{meetings.Label, i, tempt, meetings.Comment} //creates a temporary Meeting using each the meetings data
		allMeetings[i] = temp                                       //assigns temporary meeting to the key in the map
	}
	sortMeetingsByTime()

}

//function that searches allMeetings for the soonest one and returns it
func searchMeetingsForSoonest() Meeting {
	for _, v := range allMeetings {
		if time.Until(v.DateTime) > 0 && v.Index >= 0 {
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

func ShowAllMeetings(command string) string {
	initializeAllMeetings()
	_, flags, err := sFlags.CreateFlags(command)
	if err != nil {
		fmt.Println(err)
	}
	var output string
	fmt.Println(flags["-tz"])
	for k, v := range allMeetings {
		if k >= 0 {
			output += v.toStringInLoc(flags["-tz"])
			output += "Index: "
			output += strconv.Itoa(k)
			output += "\n"
		}
	}
	return output
}

//adds Meeting to the allMeetings map and to the json file
func AddMeeting(meeting string) (error, string) {
	var temp Meeting
	prevNumMeetings := len(allMeetings)
	newNumMeetings := prevNumMeetings + 1
	tokens := strings.Split(meeting, "; ") //Splits a string into the parts needed to make a meeting
	extraTokens := strings.Fields(tokens[1])
	tZone := extraTokens[len(extraTokens)-1]
	tempt, err := time.ParseInLocation(timeStamp, tokens[1], timeZones[tZone]) //turns the time string into a time object
	if err == nil {
		temp = Meeting{tokens[0], newNumMeetings, tempt, tokens[2]} //creates a temporary meet object according to the string pieces
		allMeetings[newNumMeetings-1] = temp
		sortMeetingsByTime()
	} else {
		fmt.Println(err)
	}
	sortMeetingsByTime()
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
	fmt.Println(temp.toString())
	fmt.Println("Temp Index:", temp.Index)
	fmt.Println("Deletion Index: ", idx)
	if temp.Label == "" {
		return "", errors.New("Unable to find meeting with given index")
	} else {
		delete(allMeetings, idx)
	}
	if allMeetings[idx].Label != "" {
		return "", errors.New("Unable to delete meeting")
	}
	writeMapToFile()
	sortMeetingsByTime()
	return temp.toString(), nil
}
