// Package cfg implements configuration file encoding, decoding, and storing.
package cfg

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds the available configurable options
type Config struct {
	Token   string `toml:"token"`
	GuildID string `toml:"guildID"`
	UserID  string `toml:"userID"`
}

// GetConfig parses the config.tmol file
func GetConfig() *Config {
	var conf Config
	if _, err := toml.DecodeFile("conf.toml", &conf); err != nil {
		setDefault(&conf)
		SaveConfig(&conf)
	}
	return &conf
}

// SaveConfig writes current configuration to the config.tmol file
func SaveConfig(c *Config) {
	var f *os.File
	var err error
	if f, err = os.Create("conf.toml"); err != nil {
		log.Fatalln(err)
	}
	enc := toml.NewEncoder(f)
	if err = enc.Encode(c); err != nil {
		log.Fatalln(err)
	}
}

// Config defaults go here
func setDefault(c *Config) {
	c.Token = ""
	c.GuildID = ""
}
