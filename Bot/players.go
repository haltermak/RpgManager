package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/haltermak/RpgManager/Manager"
	"github.com/haltermak/sFlags"
	"math/rand"
	"strings"
	//"time"
)

var timeSource rand.Source
var randSource *rand.Rand
var gameMap map[int]Game

type Gametype struct {
	name string
	roll func(string) (string, error)
}

/*func initGameTypes() {
	Shadowrun := Gametype{"Shadowrun", RollShadowrunDice}
	timeSource = rand.NewSource(time.Now().UnixNano())
	randSource = rand.New(timeSource)
}*/

func initGames() {
	gameMap = make(map[int]Game)
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
	gameAdmin       *Player
	id              int
	Gametype
}

/**
 * Function should take a string command and create a new Game object
 * @param command: a string with the format /newGame -gN s -gT s -nP x -p1 s -p2 s -p3... and so forth, where -gT is a string representing game type, x is the number of players, and each -px is a call to make a new player using s
 */
func NewGame(m *discordgo.MessageCreate) (*Game, error) {
	_, flags, err := sFlags.CreateFlags(m.Content)
	newgame := new(Game)
	newgame.gameAdmin = NewPlayer("/create -pN " + m.Author.Username + " -a")
	newgame.gameName = flags["-gN"]
	newgame.id = randSource.Int()
	numPlayers, err := sFlags.FlagToInt(flags, "-nP")
	pCounter := 0
	playerNames := make([]string, numPlayers)
	for _, p := range flags {
		if strings.HasPrefix(p, "-p") {
			playerNames[pCounter] = flags[p]
			pCounter++
		}
	}
	for _, name := range playerNames {
		newgame.nonAdminPlayers[name] = NewPlayer(name)
	}
	gameMap[newgame.id] = *newgame
	return newgame, err
}

/**
 * Players should contain everything needed to interact with that player, including remembering their controlled entities, if any, as well as the Game they play in
 */
type Player struct {
	pUser    discordgo.User
	name     string
	admin    bool
	entities map[string]Entity
	game     *Game
}

/**
 * Takes a command string and returns a player with hopefully the correct information
 * @param: string of format /newPlayer -pN s -a b
 */
func NewPlayer(name string) *Player {
	p := new(Player)
	p.name = name
	return p
}

func (p Player) associatePlayer(gameID int, m *discordgo.MessageCreate) {
	p.pUser = *m.Author
	bob := gameMap[gameID]
	p.game = &bob
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
