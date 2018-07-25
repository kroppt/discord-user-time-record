# Discord User Time Record Bot

[![Go Report Card](https://goreportcard.com/badge/github.com/kroppt/discord-user-time-record)](https://goreportcard.com/report/github.com/kroppt/discord-user-time-record)

Record play times for individuals via the Discord API.

## Installation

First, [install golang](https://golang.org/doc/install).

Then, download and install the package:

    $ go get -u github.com/kroppt/discord-user-time-record

Finally, run the program:

    $ discord-user-time-record

## Configuration

The file `conf.toml` will be created after inputting information to the application. This can also be edited manually.

Note: the configuration file will be created in the current directory until changed [#2](https://github.com/kroppt/discord-user-time-record/issues/2)

Example:
```
token = "zMndOe7jFLXGawdlxMOdNvXjjOce5X"
guildID = "41771983423143937"
userID = "80351110224678912"
```

## Setup

When running the program, you will be asked to supply three things:

1. Bot token

You will need to register an app in Discord and assign it a bot. Go to Discord API [My Apps](https://discordapp.com/developers/applications/me) and create an app (the app name will be the ID of the bot in the guild). You should be seeing the configuration screen for your new app.

Scroll down to the Bot section and click "Create a Bot User" and "Yes, do it!" on the popup. Under the app bot's username, the Token field should be clickable to reveal the token. Reveal it and copy the token. This is the first input to discord-user-time-record.

Now, you need to register that bot with the server you want it to run on.

Back on the app configuration page, scroll up to the "APP NAME" section. Select "Generate OAuth2 URL".

In the "SCOPES" list, "bot" is the only one that needs to be selected.

Select "COPY" and paste the copied URL into a new tab/window. Under "Add a bot to a server", select the server you want the bot to be on. The bot will show up as a user, but idefintifiable as a bot by a bot flag.

2. Guild ID

When the bot token is added to the application, it will present a list of server options. If no options appear, see above for registering a bot with a server. Select one to save it with the application.

3. User ID

When the Guild ID is added to the application, it will present a list of user options. Select one to save it with the application.

## Output

The output is very robust, containing start and end time of application execution and logging output containing game tracker changes and the time.

Afterwards, all of the apps or games that were shown on Discord will show up in a 2-column list, like so:
```
2018/04/29 04:44:21 start time: 2018-04-29 04:44:21.506234168 -0400 EDT m=+0.146273060
2018/04/29 04:44:21 tracking user "xyzabc" with ID "314159265358979323"
^C
2018/04/29 04:44:33 game tracker created: "REALLY quick game"
2018/04/29 04:44:33 stop time: 2018-04-29 04:44:33.880262979 -0400 EDT m=+12.520301901
2018/04/29 04:44:33 total running time 12.374028841s

  Final Results

|  Game               |  Time played  |
|                     |               |
|  REALLY quick game  |  12s          |
```
