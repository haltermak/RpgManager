package RpgManager

import (
	"fmt"
	"math/rand"
	"time"
	//"os"
	"errors"
	"github.com/haltermak/sFlags"
	"math"
	"sort"
	"strconv"
	"strings"
)

//@todo:  Allowing more kinds of filtering than simply low or high (keep dice), repeat rolls, exploding dice

/*Main purpose of the functions from this file are to accept a string and return the dice roll. It should use standard dice notation to roll the dice, and also accept constant modifiers,
and count successes and ones for the players. Basically rewriting sidekick*/
var timeSource rand.Source
var randSource *rand.Rand
var numDice, typeDice, operand, comparand, total, successes, keepDice int
var comparator, operator string
var compoperators = [...]string{">>", ">=", "<<", "<=", "=="}
var operators = [...]string{"+", "-", "÷", "/", "*", "x", "X"}
var minMaxFlag, roundFlag, keepFlag string

func InitDice() {
	timeSource = rand.NewSource(time.Now().UnixNano())
	randSource = rand.New(timeSource)
}

func InitDiceWithSource(s *rand.Rand) {
	randSource = s
}

//Function should be called with an argument feeding it a string to roll in the format xdy, where x is the number of dice and y is the number of sides on the dice. It can be followed with various flags,
//or comparisons and modifiers.
func RollDice(dice string) (string, error) {
	//This block of code turns a command line argument into two ints, one being the number of dice thrown, and the other being the type of dice
	clearFlags()
	rollType, diceSlice, err := inputProofer(dice)
	//fmt.Println(diceSlice)
	err = assignMeaningToDiceSlice(diceSlice, rollType)
	if err != nil {
		return "", err
	}
	if numDice > 1000 {
		return "That's too many dice", nil
	}
	if typeDice > 1000 {
		return "Golf balls are not dice", nil
	}
	if numDice+typeDice > 1000 {
		return "Fuck you", nil
	}
	//Create a source of random numbers seeded to the current time

	var results []int
	results, total, successes, err = rollDice(rollType, roundFlag)
	if err != nil {
		return "DAMNIT", err
	}
	return formatResults(results, rollType), nil
	//fmt.Println(results)
	//fmt.Println(total)
	//fmt.Println(successes)
}

/**
 * function should accept string of format "/roll xdy(==z/+o) and flags" then return the string of the dice roll, along with the successes, total, and error
 */
func RollDiceRedo(dice string) (string, int, int, error) {
	err := clearFlags()
	if err != nil {
		fmt.Println("error clearing flags")
	}
	command, flags, err := sFlags.CreateFlags(dice)
	parseFlags(flags)
	rollType, diceSlice, err := inputProofer(command)
	err = assignMeaningToDiceSlice(diceSlice, rollType)
	if err != nil {
		return "", 0, 0, err
	}
	if numDice > 1000 {
		return "That's too many dice", 0, 0, nil
	}
	if typeDice > 1000 {
		return "Golf balls are not dice", 0, 0, nil
	}
	if numDice+typeDice > 1000 {
		return "Fuck you", 0, 0, nil
	}
	var results []int
	results, total, successes, err = rollDice(rollType, roundFlag)
	if err != nil {
		return "DAMNIT", 0, 0, err
	}
	return formatResults(results, rollType), successes, total, err
}

/* function to proof the input string's format, then split it up and return an option that indicates what kind of roll should be happening */

func inputProofer(diceString string) (int, []string, error) {
	var err error
	diceString, err = stripFlags(diceString)

	containsComp := 0
	containsMod := 0
	//This block of code checks for single comparators and replaces them with doubles so that they're easier to parse
	if strings.Contains(diceString, ">") && (!strings.Contains(diceString, ">=") && !strings.Contains(diceString, ">>")) {
		diceString = strings.Replace(diceString, ">", ">>", -1)
	}
	if strings.Contains(diceString, "<") && (!strings.Contains(diceString, "<=") && !strings.Contains(diceString, "<<")) {
		diceString = strings.Replace(diceString, "<", "<<", -1)
	}
	if strings.Contains(diceString, "=") && (!strings.Contains(diceString, "<=") && !strings.Contains(diceString, "==") && !strings.Contains(diceString, ">=")) {
		diceString = strings.Replace(diceString, "=", "==", -1)
	}
	//splits the diceString into two parts, before the 'd' and after. If there are more or less than two parts, it returns zero, which is recognized as an error
	diceSlice := strings.Split(diceString, "d")
	if len(diceSlice) != 2 {
		return 0, diceSlice, err
	}

	//Checks for the presence of one of the compoperators, then splits the second part of the diceSlice accordingly.
	//Before this starts, the diceSlice should have length=2. The second element should be everything after the d in the input.
	for i := 0; i < 5; i++ {
		if strings.Contains(diceSlice[1], compoperators[i]) == true {
			temp := strings.Split(diceSlice[1], compoperators[i])
			diceSlice[1] = temp[0]
			//diceSlice[2] = compoperators[i]
			diceSlice = append(diceSlice, compoperators[i], temp[1])
			//diceSlice[3] = temp[1]
			containsComp = 1
			break
		}
	}
	if containsComp == 1 && len(diceSlice) != 4 {
		return 0, diceSlice, err
	}
	//If there were any comparisons, containsComp should be 1. That is used to adjust the destinations of the results of the search for operations. Otherwise, the same idea as the search for comparisons
	for i := 0; i < 6; i++ {
		if strings.Contains(diceSlice[1+(2*containsComp)], operators[i]) == true {
			temp := strings.Split(diceSlice[1+(2*containsComp)], operators[i])
			diceSlice[1+(2*containsComp)] = temp[0]
			//diceSlice[2+(2*containsComp)] = operators[i]
			diceSlice = append(diceSlice, operators[i], temp[1])
			//diceSlice[3+(2*containsComp)] = temp[1]
			containsMod = 1
			break
		}
	}
	if containsMod == 1 && (len(diceSlice) != 4 && len(diceSlice) != 6) {
		return 0, diceSlice, err
	}
	//diceSlice should now have everything split accordingly. Using the flags from above, the function returns an int that represents the type of die roll being carried out
	if containsComp == 0 && containsMod == 0 {
		if keepFlag == "" {
			return 1, diceSlice, err
		} else if keepFlag != "" {
			return 5, diceSlice, err
		}
	} else if containsComp == 1 && containsMod == 0 {
		if keepFlag != "" {
			return 6, diceSlice, err
		} else {
			return 2, diceSlice, err
		}
	} else if containsComp == 0 && containsMod == 1 {
		return 3, diceSlice, err
	} else {
		return 4, diceSlice, err
	}
	return 0, diceSlice, err
}

func stripFlags(s string) (string, error) {
	fullSlice := strings.Fields(s)
	flagSlice := fullSlice[1:]
	var err error
	for i, piece := range flagSlice {
		if piece == "-m" {
			minMaxFlag = flagSlice[i+1]
		} else if piece == "-r" {
			roundFlag = flagSlice[i+1]
		} else if piece == "-k" {
			interKeep := flagSlice[i+1]
			keepFlag = string(interKeep[0])
			tempString := string(interKeep[1:])
			keepDice, err = strconv.Atoi(tempString)
		}
	}
	return fullSlice[0], err
}

func parseFlags(f map[string]string) {
	minMaxFlag = f["-m"]
	roundFlag = f["-rn"]
	interKeep := f["-k"]
	keepFlag = string(interKeep[0])
	tempString := string(interKeep[1:])
	var err error
	keepDice, err = strconv.Atoi(tempString)
	if err != nil {
		fmt.Println(err)
	}
}

func clearFlags() error {
	var err error
	keepFlag = ""
	minMaxFlag = ""
	roundFlag = ""
	numDice = 0
	keepDice = 0
	comparand = 0
	operand = 0
	successes = 0
	total = 0
	return err
}

func assignMeaningToDiceSlice(diceS []string, rT int) error {
	var err error
	if rT == 1 {
		numDice, err = strconv.Atoi(diceS[0])
		typeDice, err = strconv.Atoi(diceS[1])
	} else if rT == 2 {
		numDice, err = strconv.Atoi(diceS[0])
		typeDice, err = strconv.Atoi(diceS[1])
		comparator = diceS[2]
		comparand, err = strconv.Atoi(diceS[3])
	} else if rT == 3 {
		numDice, err = strconv.Atoi(diceS[0])
		typeDice, err = strconv.Atoi(diceS[1])
		operator = diceS[2]
		operand, err = strconv.Atoi(diceS[3])
	} else if rT == 4 {
		numDice, err = strconv.Atoi(diceS[0])
		typeDice, err = strconv.Atoi(diceS[1])
		comparator = diceS[2]
		comparand, err = strconv.Atoi(diceS[3])
		operator = diceS[4]
		operand, err = strconv.Atoi(diceS[5])
	} else if rT == 5 {
		numDice, err = strconv.Atoi(diceS[0])
		typeDice, err = strconv.Atoi(diceS[1])
	} else if rT == 6 {
		numDice, err = strconv.Atoi(diceS[0])
		typeDice, err = strconv.Atoi(diceS[1])
		comparator = diceS[2]
		comparand, err = strconv.Atoi(diceS[3])
	}
	return err
}

func rollDice(rT int, rounding string) ([]int, int, int, error) {
	switch rT {
	case 0:
		return nil, 0, 0, errors.New("Dice Input not recognized")
	case 1:
		diceResults, total, err := rollBasicDice(numDice, typeDice)
		return diceResults, total, 0, err
	case 2:
		diceResults, total, err := rollBasicDice(numDice, typeDice)
		successes = determineSuccesses(diceResults, comparator, comparand)
		return diceResults, total, successes, err
	case 3:
		diceResults, total, err := rollBasicDice(numDice, typeDice)
		switch operator {
		case "+":
			total += operand
		case "-":
			total -= operand
		case "*":
			total = total * operand
		case "x":
			total = total * operand
		case "X":
			total = total * operand
		case "/":
			if rounding == "-Up" {
				total = int(math.Ceil(float64(total) / float64(operand)))
			} else if rounding == "-Dn" {
				total = int(math.Floor(float64(total) / float64(operand)))
			} else {
				total = total / operand
			}
		case "÷":
			if rounding == "-Up" {
				total = int(math.Ceil(float64(total) / float64(operand)))
			} else if rounding == "-Dn" {
				total = int(math.Floor(float64(total) / float64(operand)))
			} else {
				total = total / operand
			}
		}
		return diceResults, total, 0, err
	case 4:
		diceResults, total, err := rollBasicDice(numDice, typeDice)
		successes = determineSuccesses(diceResults, comparator, comparand)
		switch operator {
		case "+":
			total += operand
		case "-":
			total -= operand
		}
		return diceResults, total, successes, err
	case 5:
		diceResults, total, err := rollBasicDice(numDice, typeDice)
		return diceResults, total, 0, err
	case 6:
		diceResults, total, err := rollBasicDice(numDice, typeDice)
		fmt.Println(diceResults, comparator, comparand)
		successes = determineSuccesses(diceResults, comparator, comparand)
		return diceResults, total, successes, err
	default:
		return nil, 0, 0, errors.New("Invalid roll type")
	}

}

func rollBasicDice(numDice, typeDice int) ([]int, int, error) {
	var err error
	err = nil
	defer func() {
		if recover() != nil {
			err = errors.New("Error generating dice rolls")
		}
	}()
	resultsOfDice := make([]int, numDice)
	for i, _ := range resultsOfDice {
		resultsOfDice[i] = randSource.Intn(typeDice) + 1
	}
	total := 0
	for _, currentDie := range resultsOfDice {
		total += currentDie
	}
	if keepFlag == "" {
		return resultsOfDice, total, err
	} else {
		if keepFlag == "l" {
			sort.Ints(resultsOfDice)
		} else {
			sort.Sort(sort.Reverse(sort.IntSlice(resultsOfDice)))
		}
		return resultsOfDice, total, err
	}
}

func determineSuccesses(dice []int, comp string, comd int) int {
	fmt.Println(dice, comp, comd)
	success := 0
	switch comp {
	case ">=":
		for _, currentDie := range dice {
			if currentDie >= comd {
				success++
			}
		}
	case ">>":
		for _, currentDie := range dice {
			if currentDie > comd {
				success++
			}
		}
	case "<=":
		for _, currentDie := range dice {
			if currentDie <= comd {
				success++
			}
		}
	case "<<":
		for _, currentDie := range dice {
			if currentDie < comd {
				success++
			}
		}
	case "==":
		for _, currentDie := range dice {
			if currentDie == comd {
				success++
			}
		}
	}
	if keepFlag != "" {
		if keepFlag == "h" {
			if success > keepDice {
				success = keepDice
			}
		} else if keepFlag == "l" {
			if success > keepDice {
				success = keepDice
			}
		}
	}
	fmt.Println(success)
	return success
}

func formatResults(dice []int, rT int) string {
	output := ""
	switch rT {
	case 1:
		output += "("
		for _, die := range dice {
			output += strconv.Itoa(die)
			output += "+"
		}
		output = strings.TrimSuffix(output, "+")
		output += ")"
		output += "="
		output += strconv.Itoa(total)
	case 2:
		output += "["

		for _, die := range dice {
			var numbString string
			switch comparator {
			case ">=":
				if die < comparand {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
				}
			case ">>":
				if die <= comparand {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
				}
			case "<=":
				if die > comparand {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
				}
			case "<<":
				if die >= comparand {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
				}
			case "==":
				if die != comparand {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
				}
			}
			output += numbString
			output += ","
		}
		output = strings.TrimSuffix(output, ",")
		output += "]"
		output += " giving "
		output += strconv.Itoa(successes)
		output += " successes"

	case 3:
		output += "("
		for _, die := range dice {
			output += strconv.Itoa(die)
			output += "+"
		}
		output = strings.TrimSuffix(output, "+")
		output += ")"
		output += operator
		output += strconv.Itoa(operand)
		output += "="
		output += strconv.Itoa(total)
	case 5:
		output += "["
		for i, die := range dice {
			if keepFlag == "h" {
				if i < (keepDice) {
					output += strconv.Itoa(die)
				} else {
					output += "~~" + strconv.Itoa(die) + "~~"
				}
			} else {
				if i < (keepDice) {
					output += strconv.Itoa(die)
				} else {
					output += "~~" + strconv.Itoa(die) + "~~"
				}
			}
			output += ","
		}
		output = strings.TrimSuffix(output, " ")
		output += "]"
	case 6:
		output += "["
		successCounter := 0
		for _, die := range dice {
			var numbString string
			switch comparator {
			case ">=":
				if die < comparand || successCounter >= keepDice {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
					successCounter++
				}
			case ">>":
				if die <= comparand || successCounter >= keepDice {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
					successCounter++
				}
			case "<=":
				if die > comparand || successCounter >= keepDice {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
					successCounter++
				}
			case "<<":
				if die >= comparand || successCounter >= keepDice {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
					successCounter++
				}
			case "==":
				if die != comparand || successCounter >= keepDice {
					numbString = "~~" + strconv.Itoa(die) + "~~"
				} else {
					numbString = strconv.Itoa(die)
					successCounter++
				}
			}
			output += numbString
			output += ","
		}
		output = strings.TrimSuffix(output, ",")
		output += "]"
		output += " giving "
		output += strconv.Itoa(successes)
		output += " successes"
	}

	afterEquals := ""
	if strings.Contains(output, "=") {
		temp := strings.Split(output, "=")
		output = temp[0]
		afterEquals = "=" + temp[1]
	}
	if strings.Contains(output, "giving") {
		temp := strings.Split(output, "giving")
		temp[0] += "giving"
		output = temp[0]
		afterEquals = temp[1]
	}
	switch minMaxFlag {
	case "mM":
		output = strings.Replace(output, "1", "**1**", -1)
		output = strings.Replace(output, strconv.Itoa(typeDice), "**"+strconv.Itoa(typeDice)+"**", -1)
	case "m":
		output = strings.Replace(output, "1", "**1**", -1)
	case "M":
		output = strings.Replace(output, strconv.Itoa(typeDice), "**"+strconv.Itoa(typeDice)+"**", -1)
	}
	clearFlags()
	return output + afterEquals
}
