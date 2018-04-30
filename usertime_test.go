package main

import (
	"reflect"
	"testing"

	dg "github.com/bwmarrin/discordgo"
	"github.com/kroppt/discord-user-time-record/cfg"
)

func Test_printResults(t *testing.T) {
	trackList = &trackerList{}
	trackList.Trackers = append(trackList.Trackers,
		&gameTracker{dg.Game{Name: "game1"}, 1000})
	trackList.Trackers = append(trackList.Trackers,
		&gameTracker{dg.Game{Name: "game2"}, 1234})
	tests := []struct {
		name string
	}{
		{name: "normal use"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printResults()
		})
	}
}

func Test_updatePlayTime(t *testing.T) {
	trackList = &trackerList{Trackers: []*gameTracker{
		&gameTracker{dg.Game{Name: "game1"}, 1000},
	}}
	trackList.Trackers = append(trackList.Trackers)
	type args struct {
		p int64
		g *dg.Game
	}
	tests := []struct {
		name string
		args args
		want []*gameTracker
	}{
		{
			"update existing tracker",
			args{500, &dg.Game{Name: "game1"}},
			[]*gameTracker{
				&gameTracker{dg.Game{Name: "game1"}, 1500},
			},
		},
		{
			"create new tracker",
			args{500, &dg.Game{Name: "game2"}},
			[]*gameTracker{
				&gameTracker{dg.Game{Name: "game1"}, 1500},
				&gameTracker{dg.Game{Name: "game2"}, 500},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatePlayTime(tt.args.p, tt.args.g)
			for i, got := range trackList.Trackers {
				if !reflect.DeepEqual(got, tt.want[i]) {
					t.Errorf("trackList.Trackers[%d] = %v, want[%d] %v",
						i, got, i, tt.want[i])
				}
			}
			lenGot := len(trackList.Trackers)
			lenWant := len(tt.want)
			if lenGot != lenWant {
				t.Errorf("len(trackList.Trackers) = %v, len(want) %v", lenGot, lenWant)
			}
		})
	}
}

func Test_guildCreate(t *testing.T) {
	trackList = &trackerList{}
	conf = &cfg.Config{UserID: "test UserID"}
	type args struct {
		s     *dg.Session
		event *dg.GuildCreate
	}
	tests := []struct {
		name string
		args args
		want dg.Presence
	}{
		{
			"guild unavailable",
			args{nil, &dg.GuildCreate{Guild: &dg.Guild{Unavailable: true}}},
			dg.Presence{},
		},
		{
			"not playing game",
			args{nil, &dg.GuildCreate{Guild: &dg.Guild{
				Unavailable: false,
				Presences: []*dg.Presence{
					&dg.Presence{User: &dg.User{ID: "invalid UserID"}},
				},
			}}},
			dg.Presence{},
		},
		{
			"playing game",
			args{nil, &dg.GuildCreate{Guild: &dg.Guild{
				Unavailable: false,
				Presences: []*dg.Presence{
					&dg.Presence{User: &dg.User{ID: "test UserID"}},
				},
			}}},
			dg.Presence{User: &dg.User{ID: "test UserID"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guildCreate(tt.args.s, tt.args.event)
			if got := trackList.LastPresence; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trackList.LastPresence = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_presenceUpdate(t *testing.T) {
	trackList = &trackerList{}
	conf = &cfg.Config{UserID: "test UserID"}
	type args struct {
		s     *dg.Session
		event *dg.PresenceUpdate
	}
	tests := []struct {
		name string
		args args
		want dg.Presence
	}{
		{
			"not playing game, invalid UserID",
			args{nil, &dg.PresenceUpdate{
				Presence: dg.Presence{User: &dg.User{ID: "invalid UserID"}},
			}},
			dg.Presence{},
		},
		{
			"not playing game, valid UserID",
			args{nil, &dg.PresenceUpdate{
				Presence: dg.Presence{User: &dg.User{ID: "test UserID"}},
			}},
			dg.Presence{User: &dg.User{ID: "test UserID"}},
		},
		{
			"from not playing game to playing game",
			args{nil, &dg.PresenceUpdate{
				Presence: dg.Presence{
					User: &dg.User{ID: "test UserID"},
					Game: &dg.Game{Name: "game1"},
				},
			}},
			dg.Presence{
				User: &dg.User{ID: "test UserID"},
				Game: &dg.Game{Name: "game1"},
			},
		},
		{
			"playing game, invalid UserID",
			args{nil, &dg.PresenceUpdate{
				Presence: dg.Presence{
					User: &dg.User{ID: "invalid UserID"},
					Game: &dg.Game{Name: "game1"},
				},
			}},
			dg.Presence{
				User: &dg.User{ID: "test UserID"},
				Game: &dg.Game{Name: "game1"},
			},
		},
		{
			"from playing game to playing different game",
			args{nil, &dg.PresenceUpdate{
				Presence: dg.Presence{
					User: &dg.User{ID: "test UserID"},
					Game: &dg.Game{Name: "game2"},
				},
			}},
			dg.Presence{
				User: &dg.User{ID: "test UserID"},
				Game: &dg.Game{Name: "game2"},
			},
		},
		{
			"from playing game to not playing game",
			args{nil, &dg.PresenceUpdate{
				Presence: dg.Presence{
					User: &dg.User{ID: "test UserID"},
				},
			}},
			dg.Presence{
				User: &dg.User{ID: "test UserID"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenceUpdate(tt.args.s, tt.args.event)
			if got := trackList.LastPresence; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trackList.LastPresence = %v, want %v", got, tt.want)
			}
		})
	}
}
