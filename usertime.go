package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	cfg "github.com/kroppt/discord-user-time-record/cfg"

	"github.com/bwmarrin/discordgo"
	cli "github.com/kroppt/climenu"
)

var conf *cfg.Config

type trackerList struct {
	sync.Mutex
	LastPresence discordgo.Presence
	LastTime     int64
	Trackers     []*gameTracker
}

type gameTracker struct {
	Game     discordgo.Game
	PlayTime int64 // in seconds
}

var trackList *trackerList

func main() {
	conf = cfg.GetConfig()
	var err error

	if conf.Token == "" {
		tokQuery := cli.GetText("Enter bot token", "")
		if tokQuery == "" {
			log.Fatalln("Empty bot token entered")
		}
		conf.Token = tokQuery
	}
	var dg *discordgo.Session
	if dg, err = discordgo.New("Bot " + conf.Token); err != nil {
		log.Fatalln(err)
	}
	cfg.SaveConfig(conf)

	if conf.GuildID == "" {
		var guilds []*discordgo.UserGuild
		if guilds, err = dg.UserGuilds(15, "", ""); err != nil {
			log.Fatalln(err)
		}
		guildmenu := cli.NewButtonMenu("", "Select guild channel")
		for _, g := range guilds {
			guildmenu.AddMenuItem(g.Name, g.ID)
		}
		var gID string
		var esc bool
		if gID, esc = guildmenu.Run(); esc != false {
			fmt.Println("Guild selection escaped")
			return
		}
		conf.GuildID = gID
		cfg.SaveConfig(conf)
	}

	if conf.UserID == "" {
		var gMembers []*discordgo.Member
		limit := 9
		if gMembers, err = dg.GuildMembers(conf.GuildID, "", limit); err != nil {
			log.Fatalln(err)
		}
		var isMore = false
		for {
			var lastID string
			memmenu := cli.NewButtonMenu("", "Select user")

			for _, gMem := range gMembers {
				memmenu.AddMenuItem(gMem.User.Username, gMem.User.ID)
				lastID = gMem.User.ID
			}

			if len(gMembers) == limit {
				memmenu.AddMenuItem("More members", "more")
			} else if isMore {
				memmenu.AddMenuItem("Reset list", "reset")
			}

			var selID string
			var esc bool
			if selID, esc = memmenu.Run(); esc != false {
				fmt.Println("Member selection escaped")
				return
			}
			if selID == "reset" {
				gMembers, err = dg.GuildMembers(conf.GuildID, "", limit)
				isMore = false
			} else if selID == "more" {
				gMembers, err = dg.GuildMembers(conf.GuildID, lastID, limit)
				isMore = true
			} else {
				conf.UserID = selID
				cfg.SaveConfig(conf)
				break
			}
		}
	}

	trackList = &trackerList{}

	dg.AddHandler(guildCreate)
	dg.AddHandler(presenceUpdate)

	user, err := dg.User(conf.UserID)
	if err != nil {
		log.Fatalf("Failed to fetch User with UserID \"%s\"\n", conf.UserID)
	}
	// Start time
	starttime := time.Now()
	log.Printf("start time: %v\n", starttime)
	log.Printf("tracking user \"%s\" with ID \"%s\"\n", user.Username, user.ID)

	if err = dg.Open(); err != nil {
		log.Fatalln(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println()

	// Cleanly close down the Discord session.
	dg.Close()

	stoptime := time.Now()
	leftOverTime := stoptime.Unix() - trackList.LastTime

	trackList.Lock()
	var def int64
	lastGame := trackList.LastPresence.Game
	if trackList.LastTime != def && lastGame != nil {
		found := false
		for _, gt := range trackList.Trackers {
			if gt.Game.Name == trackList.LastPresence.Game.Name {
				found = true
				gt.PlayTime += leftOverTime
				log.Printf("game tracker created: \"%s\"\n", gt.Game.Name)
				break
			}
		}
		if !found {
			trackList.Trackers = append(trackList.Trackers, &gameTracker{*lastGame, leftOverTime})
			log.Printf("game tracker created: \"%s\"\n", lastGame.Name)
		}
	}
	trackList.Unlock()

	// Stop time
	log.Printf("stop time: %v\n", stoptime)
	// Stop time minus start time
	log.Printf("total running time %v\n", stoptime.Sub(starttime))

	printResults()
}

func printResults() {
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug|tabwriter.DiscardEmptyColumns)
	fmt.Printf("\n  Final Results\n\n")
	fmt.Fprintln(w, "\t  Game  \t  Time played  \t")
	trackList.Lock()
	for _, game := range trackList.Trackers {
		fmt.Fprintf(w, "\t\t\t\n")
		fmt.Fprintf(w, "\t  %s  \t  %v  \t\n", game.Game.Name, time.Duration(game.PlayTime)*time.Second)
	}
	trackList.Unlock()
	w.Flush()
	fmt.Println()
}

func updatePlayTime(p int64, g *discordgo.Game) {
	for _, t := range trackList.Trackers {
		if t.Game.Name == g.Name {
			t.PlayTime += p
			// Target hit
			log.Printf("game tracker updated: \"%s\"\n", g.Name)
			return
		}
	}
	log.Printf("game tracker created: \"%s\"\n", g.Name)
	trackList.Trackers = append(trackList.Trackers, &gameTracker{*g, p})
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined, including when the bot starts up.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	startTime := time.Now().Unix()
	if event.Guild.Unavailable {
		return
	}
	for _, p := range event.Presences {
		if p.User.ID == conf.UserID {
			// Target acquired
			trackList.Lock()
			trackList.LastPresence = *p
			trackList.LastTime = startTime
			trackList.Unlock()
			break
		}
	}
	// Target evaded
}

func presenceUpdate(s *discordgo.Session, event *discordgo.PresenceUpdate) {
	updateTime := time.Now().Unix()
	if event.User.ID == conf.UserID {
		// Target acquired
		trackList.Lock()
		defer trackList.Unlock()
		last := trackList.LastPresence
		trackList.LastPresence = event.Presence
		lastTime := trackList.LastTime
		trackList.LastTime = updateTime
		canUpdate := last.Game != nil
		if canUpdate && (event.Game == nil || event.Game.Name != last.Game.Name) {
			newTime := updateTime - lastTime
			// Fire when ready
			updatePlayTime(newTime, last.Game)
		}
	}
}
