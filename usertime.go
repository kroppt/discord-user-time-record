package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	cli "github.com/kroppt/climenu"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var token string
	var err error
	var f *os.File
	dat, err := ioutil.ReadFile("apikey.txt")
	token = string(dat)
	if f, err = os.OpenFile("apikey.txt", os.O_RDWR, os.ModeDir); err != nil {
		if f, err = os.Create("apikey.txt"); err != nil {
			log.Fatalln(err)
		}
	}
	var dg *discordgo.Session
	if token == "" {
		user := cli.GetText("Enter Discord username", "")
		pass := cli.GetText("Enter Discord password", "")
		if dg, err = discordgo.New([]string{user, pass}); err != nil {
			log.Fatalln(err)
		}
		token = dg.Token
	} else if dg, err = discordgo.New(token); err != nil {
		log.Fatalln(err)
	}
	if _, err = f.WriteString(token); err != nil {
		log.Fatalln(err)
	}
	var guilds []*discordgo.UserGuild
	if guilds, err = dg.UserGuilds(15, "", ""); err != nil {
		log.Fatalln(err)
	}
	chanmenu := cli.NewButtonMenu("", "Select user channel")
	for _, g := range guilds {
		chanmenu.AddMenuItem(g.Name, g.ID)
	}
	var selID string
	var esc bool
	if selID, esc = chanmenu.Run(); esc != false {
		log.Fatalln("Channel selection escaped")
	}
	fmt.Println(selID)
}
