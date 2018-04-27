package main

import (
	"fmt"
	"log"

	cfg "github.com/kroppt/discord-user-time-record/cfg"

	"github.com/bwmarrin/discordgo"
	cli "github.com/kroppt/climenu"
)

var conf *cfg.Config

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

		}
	}
}
