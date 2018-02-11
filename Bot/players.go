package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/haltermak/RpgManager/Manager"
	"github.com/haltermak/sFlags"
	"math/rand"
	"strings"
	"time"
)

var timeSource rand.Source
var randSource *rand.Rand

type Gametype struct {
	name string
	roll func(string) (string, error)
}

func initGameTypes() {
	Shadowrun := Gametype{"Shadowrun", RollShadowrunDice}
	timeSource = rand.NewSource(time.Now().UnixNano())
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
 * Games should contain the ruleset, the players, the game master, and the guild game is played in.
 */
type Game struct {
	nonAdminPlayers map[string]*Player
	gameName        string
	guild           discordgo.Guild
	gameAdmin       Player
	id              int
	Gametype
}

/**
 * Function should take a string command and create a new Game object
 * @param command: a string with the format /newGame -gN s -gT s -nP x -p1 s -p2 s -p3... and so forth, where -gT is a string representing game type, x is the number of players, and each -px is a call to make a new player using s
 */
func NewGame(command string) *Game {
	_, flags, err := sFlags.CreateFlags(command)
	newgame := new(Game)
	newgame.gameAdmin = newPlayer()
	newgame.gameName = flags["-gN"]
	newgame.id = randSource.Int()
	numPlayers = sFlags.FlagToInt(flags, "-nP")
	pCounter := 0
	players := make([]string, 0, numPlayers)
	for k, p := range flags {
		if strings.HasPrefix(k, "-p") {
			players[pCounter] = p
		}
	}
	newgame.nonAdminPlayers
}

/**
 * Players should contain everything needed to interact with that player, including remembering their controlled entities, if any, as well as the Game they play in
 */
type Player struct {
	pUser    discordgo.User
	name     string
	admin    bool
	entities map[string]Entity
}

/**
 * Takes a command string and returns a player with hopefully the correct information
 * @param: string of format /newPlayer -pN s -a b
 */
func NewPlayer(command string) *Player {
	_, flags, err := sFlags.CreateFlags(command)
	p := new(Player)
	p.name = flags["-pN"]
	p.admin, err = sFlags.FlagToBool(flags, ["-a"])
	return *p
}

func (p Player) associatePlayer (m *discordgo.MessageCreate) {
	p.pUser = m.Author
}

/**
 * Entity can track a variety of stats based on the game rules, so they all implement the entity interface.
 */

type Entity interface {
	getName() string
	getPlayer() *Player
	newEntity() *Entity
}

type ShadowrunEntity struct {
	name   string
	player *Player
	Entity
}

func (s ShadowrunEntity) getName() string {
	return s.name
}

func (s ShadowrunEntity) getPlayer() *Player {
	return s.player
}
