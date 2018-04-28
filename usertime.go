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
	LastFetch    int64
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

	starttime := time.Now()
	log.Printf("start time: %v\n", starttime)

	if err = dg.Open(); err != nil {
		log.Fatalln(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println()

	// Cleanly close down the Discord session.
	stoptime := time.Now()
	log.Printf("stop time: %v\n", stoptime)
	log.Printf("total running time %v\n", stoptime.Sub(starttime))
	dg.Close()
	leftOverTime := time.Now().Unix() - trackList.LastFetch

	trackList.Lock()
	var def int64
	// Start time
	if trackList.LastFetch != def {
		// End time minus start time
		found := false
		for _, gt := range trackList.Trackers {
			if gt.Game.Name == trackList.LastPresence.Game.Name {
				found = true
				gt.PlayTime += leftOverTime
				break
			}
		}
		if !found {
			trackList.Trackers = append(trackList.Trackers, &gameTracker{*trackList.LastPresence.Game, leftOverTime})
		}
	}
	trackList.Unlock()

	printResults()
}

func printResults() {
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.Debug)
	fmt.Println("\nFinal Results")
	fmt.Fprintln(w, "Game\tTime played\t")
	for _, game := range trackList.Trackers {
		fmt.Fprintf(w, "%s\t%v\t\n", game.Game.Name, time.Duration(game.PlayTime)*time.Second)
	}
	w.Flush()
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined, including when the bot starts up.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	var userID string
	for _, p := range event.Presences {
		if p.User.ID == conf.UserID {
			// Target acquired
			trackList.Lock()
			trackList.LastPresence = *p
			trackList.LastFetch = time.Now().Unix()
			userID = p.User.ID
			trackList.Unlock()
			if p.Game != nil {
				log.Printf("handler guildCreate: game name %s\n", p.Game.Name)
			}
			break
		}
	}
	// Target evaded
	if userID != conf.UserID {
		log.Println("handler guildCreate: user not online")
	}
}

func presenceUpdate(s *discordgo.Session, event *discordgo.PresenceUpdate) {
	if event.User.ID == conf.UserID {
		// Target acquired
		trackList.Lock()
		defer trackList.Unlock()

		last := trackList.LastPresence
		if last.Game != nil {
			log.Printf("handler presenseUpdate: last game name %s\n", last.Game.Name)
		}
		trackList.LastPresence = event.Presence
		trackList.LastFetch = time.Now().Unix()
		if event.Presence.Game != nil {
			log.Printf("handler presenseUpdate: updated game name %s\n", event.Presence.Game.Name)
		}
		if event.Game == nil || event.Game.Name == "" || (last.Game != nil && event.Game.Name == last.Game.Name) {
			// if no game object, no game name, or same name as last status (no change)
			// Target evaded
			return
		}
		var newTime int64
		if last.Game != nil {
			newTime = last.Game.TimeStamps.StartTimestamp - event.Game.TimeStamps.StartTimestamp
		}
		for _, t := range trackList.Trackers {
			if t.Game.Name == event.Game.Name {
				t.PlayTime += newTime
				return
			}
		}
		trackList.Trackers = append(trackList.Trackers, &gameTracker{*event.Game, newTime})

	}
	// Target evaded
}
