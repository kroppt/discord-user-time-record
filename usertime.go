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
var starttime time.Time

type trackerList struct {
	sync.Mutex
	LastPresence *discordgo.Presence
	Trackers     []*gameTracker
}

type gameTracker struct {
	Game     *discordgo.Game
	PlayTime int64 // in seconds
}

var trackList *trackerList

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

	starttime = time.Now()

	if err = dg.Open(); err != nil {
		log.Fatalln(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
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

	for _, p := range event.Presences {
		if p.User.ID == conf.UserID {
			// Target acquired
			trackList.Lock()
			trackList.LastPresence = p
			trackList.Unlock()
			if p.Game != nil {
				log.Printf("game name %s\n", p.Game.Name)
			}
			break
		}
		// Target evaded
	}
	if trackList.LastPresence == nil {
		log.Println("user not online")
	}
}

func presenceUpdate(s *discordgo.Session, event *discordgo.PresenceUpdate) {
	if event.User.ID == conf.UserID {
		// Target acquired

		trackList.Lock()
		defer trackList.Unlock()

		last := trackList.LastPresence
		if last != nil && last.Game != nil {
			log.Printf("last game name %s\n", last.Game.Name)
		}
		trackList.LastPresence = &event.Presence
		if event.Presence.Game != nil {
			log.Printf("updated game name %s\n", event.Presence.Game.Name)
		}
		if event.Game == nil || event.Game.Name == "" || ((last != nil && last.Game != nil) && event.Game.Name == last.Game.Name) {
			// if no game object, no game name, or same name as last status (no change)
			return
		}
		var newTime int64
		if last != nil && last.Game != nil {
			newTime = last.Game.TimeStamps.StartTimestamp - event.Game.TimeStamps.StartTimestamp
		}
		for _, t := range trackList.Trackers {
			if t.Game != nil && (t.Game.Name == event.Game.Name) {
				t.PlayTime += newTime
				return
			}
		}
		trackList.Trackers = append(trackList.Trackers, &gameTracker{event.Game, newTime})

	}
	// Target evaded
}
