package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/haltermak/RpgManager/Manager"
	"github.com/haltermak/sFlags"
	//"io/ioutil"
	"math/rand"
	//"os"
	"strconv"
	"strings"
	"time"
)

var timeSource rand.Source
var randSource *rand.Rand
var gameMap map[int]Game

type Gametype struct {
	Name string
	roll func(string) (string, error)
}

/*func initGameTypes() {
	Shadowrun := Gametype{"Shadowrun", RollShadowrunDice}
	timeSource = rand.NewSource(time.Now().UnixNano())
	randSource = rand.New(timeSource)
}*/

func initGames() {
	gameMap = make(map[int]Game)
	timeSource = rand.NewSource(time.Now().Unix())
	randSource = rand.New(timeSource)
}

/**
 * @param  {c should be a string of format "/sr -p x -l y"}
 * @return {string} should be the result of a shadowrun type roll, including a count of hits, 1s, whether or not there was a glitch or critical glitch
 */
func RollShadowrunDice(c string) (string, error) {
	_, flags, err := sFlags.CreateFlags(c)
	pool, err := sFlags.FlagToInt(flags, "-p")
	if err != nil {
		fmt.Println(err)
	}
	limit, err := sFlags.FlagToInt(flags, "-l")
	if err != nil {
		fmt.Println(err)
	}
	command := fmt.Sprintf("%dd6>=5 -k h%d", pool, limit)
	result, successes, _, err := RpgManager.RollDiceRedo(command)
	if strings.Count(result, "1") > (pool/2) && successes == 0 {
		result += "\nCritical glitch!"
	} else if strings.Count(result, "1") > (pool / 2) {
		result += "\nGlitch!"
	}
	return result, err
}

/**
 * Games should contain the ruleset, the players, the Game master, and the guild Game is played in.
 */
type Game struct {
	NonAdminPlayers map[string]*Player
	GameName        string
	Guild           *discordgo.Guild
	GameAdmin       *Player
	Id              int
	JoinedPlayers   int
}

func (g Game) addPlayer(m *discordgo.MessageCreate) (string, error) {
	_, flags, err := sFlags.CreateFlags(m.Content)
	if err != nil {
		fmt.Println(err)
	}
	plName := flags["-pN"]
	fmt.Println(g)
	response, err := g.NonAdminPlayers[plName].AssociatePlayer(m)
	return response, err
}

/**
 * Function should take a string command and create a new Game object
 * @param command: a string with the format /newGame -gN s -gT s -nP x -p1 s -p2 s -p3... and so forth, where -gT is a string representing Game type, x is the number of players, and each -px is a call to make a new player using s
 */
func NewGame(m *discordgo.MessageCreate) (*Game, string, error) {
	_, flags, err := sFlags.CreateFlags(m.Content)
	newgame := new(Game)
	newgame.NonAdminPlayers = make(map[string]*Player)
	newgame.GameAdmin = NewPlayer("/create -pN " + m.Author.Username + " -a")
	newgame.GameName = flags["-gN"]
	newgame.Id = randSource.Intn(1)
	numPlayers, err := sFlags.FlagToInt(flags, "-nP")
	pCounter := 0
	playerNames := make([]string, numPlayers)
	fmt.Println(flags)
	for k, p := range flags {
		if strings.HasPrefix(k, "-p") {
			playerNames[pCounter] = p
			pCounter++
		}
	}
	fmt.Println(playerNames)
	for _, Name := range playerNames {
		newgame.NonAdminPlayers[Name] = NewPlayer(Name)
	}
	return newgame, strconv.Itoa(newgame.Id), err
}

/**
 * Players should contain everything needed to interact with that player, including remembering their controlled entities, if any, as well as the Game they play in
 */
type Player struct {
	PUser    discordgo.User
	Name     string
	Admin    bool
	Entities map[string]Entity
	Game     *Game
	Index    int
}

/**
 * Takes a command string and returns a player with hopefully the correct information
 * @param: string of format /newPlayer -pN s -a b
 */
func NewPlayer(Name string) *Player {
	p := new(Player)
	p.Name = Name
	return p
}

func (p Player) AssociatePlayer(m *discordgo.MessageCreate) (string, error) {
	_, flags, err := sFlags.CreateFlags(m.Content)
	if err != nil {
		fmt.Println(err)
	}
	gameID := flags["-gID"]
	p.PUser = *m.Author
	gameIDint, err := strconv.Atoi(gameID)
	if err != nil {
		fmt.Println(err)
	}
	bob := gameMap[gameIDint]
	p.Game = &bob
	return m.Author.Username + "has joined Game " + strconv.Itoa(p.Game.Id), err
}

/**
 * Entity can track a variety of stats based on the Game rules, so they all implement the entity interface.
 */

type Entity interface {
	getName() string
	getPlayer() *Player
	newEntity() *Entity
}

type ShadowrunEntity struct {
	Name   string
	player *Player
	Entity
}

func (s ShadowrunEntity) getName() string {
	return s.Name
}

func (s ShadowrunEntity) getPlayer() *Player {
	return s.player
}
