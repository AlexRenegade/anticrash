package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

var (
	version      = "1.2.3"
	botID        string
	urlRegex     = regexp.MustCompile(`(https?)://[A-Za-z0-9/._-]+`)
	gfyNameRegex = regexp.MustCompile(`/([A-Za-z0-9]+)`)
	ffProbePath  string
	cfg          config
)

type config struct {
	Prefix         string
	Token          string
	ReplyToMessage bool
	LogChannel     string
}

func init() {
	var err error

	if len(os.Args) > 1 {
		_, err = toml.DecodeFile(os.Args[1], &cfg)
	} else {
		_, err = toml.DecodeFile("config.toml", &cfg)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "💔 Couldn't read from config file:", err)
		os.Exit(1)
	}

	if cfg.Token == "" {
		fmt.Fprintln(os.Stderr, "💔 Missing Bot Token, exiting")
		os.Exit(1)
	}

	if cfg.Prefix == "" {
		fmt.Fprintln(os.Stderr, "💔 Missing command prefix, exiting")
		os.Exit(1)
	}

	ffprobe, err := exec.LookPath("ffprobe")

	if err != nil {
		fmt.Fprintln(os.Stderr, "💔 Couldn't get ffprobe path:", err)
		os.Exit(1)
	}

	ffProbePath = ffprobe
}

func main() {
	bot, err := discordgo.New("Bot " + cfg.Token)

	if err != nil {
		fmt.Fprintln(os.Stderr, "💔 Couldn't create Discord session:", err)
		os.Exit(1)
	}

	fmt.Printf("\n")

	bot.AddHandler(ready)
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)
	bot.AddHandler(guildCreate)
	bot.AddHandler(guildUpdate)
	bot.AddHandler(guildDelete)

	err = bot.Open()

	if err != nil {
		fmt.Fprintln(os.Stderr, "💔 Couldn't establish WebSocket connection:", err)
		os.Exit(1)
	}

	fmt.Println("👑 Bot running, press CTRL+C to exit")
	syscalls := make(chan os.Signal, 1)
	signal.Notify(syscalls, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt, os.Kill)
	fmt.Printf("🔺 Signal `%v` detected, disconnecting bot and exiting\n\n", <-syscalls)

	_ = bot.Close()
}

func ready(session *discordgo.Session, event *discordgo.Ready) {
	botID = event.User.ID
	fmt.Println("👤 Logged in as", event.User.String())

	_ = session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: fmt.Sprintf("+help | Protecting %d guilds", len(session.State.Guilds)),
				Type: discordgo.ActivityTypeGame,
				URL:  "",
			},
		},
		AFK:    false,
		Status: "dnd",
	})
}

func guildCreate(session *discordgo.Session, _ *discordgo.GuildCreate) {
	_ = session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: fmt.Sprintf("+help | Protecting %d guilds", len(session.State.Guilds)),
				Type: discordgo.ActivityTypeGame,
				URL:  "",
			},
		},
		AFK:    false,
		Status: "dnd",
	})
}

func guildUpdate(session *discordgo.Session, _ *discordgo.GuildUpdate) {
	_ = session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: fmt.Sprintf("+help | Protecting %d guilds", len(session.State.Guilds)),
				Type: discordgo.ActivityTypeGame,
				URL:  "",
			},
		},
		AFK:    false,
		Status: "dnd",
	})
}

func guildDelete(session *discordgo.Session, _ *discordgo.GuildDelete) {
	_ = session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: fmt.Sprintf("+help | Protecting %d guilds", len(session.State.Guilds)),
				Type: discordgo.ActivityTypeGame,
				URL:  "",
			},
		},
		AFK:    false,
		Status: "dnd",
	})
}

func checkAdminPermissions(session *discordgo.Session, event *discordgo.MessageCreate) bool {
	permissions, err := session.State.MessagePermissions(event.Message)

	if err != nil {
		return false
	}

	if (permissions & discordgo.PermissionAdministrator) != discordgo.PermissionAdministrator {
		return false
	}

	return true
}

func messageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.ID == botID {
		return
	}

	if len(event.Mentions) == 1 {
		if event.Mentions[0].ID == botID && checkAdminPermissions(session, event) {
			_, _ = session.ChannelMessageSendReply(event.ChannelID, fmt.Sprintf("Hello, I'm AntiCrash. Use `%shelp` to view available commands. Please note, commands are Administrator only!", cfg.Prefix), event.Reference())
			return
		}
	}

	if strings.HasPrefix(event.Content, cfg.Prefix) {
		if !checkAdminPermissions(session, event) {
			return
		}

		commandParts := strings.Split(event.Content, " ")
		command := strings.TrimPrefix(commandParts[0], cfg.Prefix)

		switch command {
		case "help":
			logsFieldValue := "Configure which channel the bot should log to\nCurrently "

			if cfg.LogChannel == "" {
				logsFieldValue += "disabled"
			} else {
				logsFieldValue += fmt.Sprintf("<#%s>", cfg.LogChannel)
			}

			_, _ = session.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
				Title: "AntiCrash - A bot to detect and delete crash files",
				Description: "Invite me to your own server " +
					"[here](https://discord.com/oauth2/authorize?client_id=839625900860899368&permissions=93184&scope=bot) " +
					"or [check out the repo](https://gitlab.com/honour/anticrash)" +
					"\n\n**Commands:**",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  fmt.Sprintf("`%shelp`", cfg.Prefix),
						Value: "Show this message",
					},
					{
						Name:  fmt.Sprintf("`%scheck`", cfg.Prefix),
						Value: fmt.Sprintf("Check whether a URL leads to a crash file\nUsage: `%scheck url`", cfg.Prefix),
					},
				},
				Author: &discordgo.MessageEmbedAuthor{
					Name: "Version " + version,
				},
				Color: 15761961,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "AntiCrash by xela#0049",
					IconURL: "https://cdn.discordapp.com/app-icons/839625900860899368/37c01e6b5336f22339a31a1ccda644de.png",
				},
			})

			return
		case "check":
			_ = session.ChannelMessageDelete(event.ChannelID, event.Message.ID)
			message, err := session.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
				Title: "Please wait, checking URL",
				Color: 15761961,
				Author: &discordgo.MessageEmbedAuthor{
					Name: "Version " + version,
				},
			})

			cleanEmbed := discordgo.MessageEmbed{
				Title: "No crash file(s) found",
				Color: 7143168,
				Author: &discordgo.MessageEmbedAuthor{
					Name: "Version " + version,
				},
			}

			crashEmbed := discordgo.MessageEmbed{
				Title: "Crash file detected!",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "URL",
						Value: "",
					},
				},
				Color: 15275537,
				Author: &discordgo.MessageEmbedAuthor{
					Name: "Version " + version,
				},
			}

			urls := getURLsFromMessage(event.Content, event.Attachments)

			for _, url := range urls {
				if handleCrashURL(session, event.Message, url) {
					if err != nil {
						_, _ = session.ChannelMessageSend(event.ChannelID, fmt.Sprintf("`%s` was a crash file", url))
					} else {
						crashEmbed.Fields[0].Value = fmt.Sprintf("`%s`", url)
						_, _ = session.ChannelMessageEditEmbed(message.ChannelID, message.ID, &crashEmbed)
					}
					return
				}
			}

			if err != nil {
				_, _ = session.ChannelMessageSend(event.ChannelID, "No crash file(s) found")
			} else {
				_, _ = session.ChannelMessageEditEmbed(message.ChannelID, message.ID, &cleanEmbed)
			}
			return
		}
	}

	urls := getURLsFromMessage(event.Content, event.Attachments)

	for _, url := range urls {
		if handleCrashURL(session, event.Message, url) {
			return
		}
	}
}

func messageUpdate(session *discordgo.Session, event *discordgo.MessageUpdate) {
	if event.Author != nil {
		if event.Author.ID == botID {
			return
		}
	}

	urls := getURLsFromMessage(event.Content, event.Attachments)

	for _, url := range urls {
		if handleCrashURL(session, event.Message, url) {
			return
		}
	}
}

func handleCrashURL(session *discordgo.Session, message *discordgo.Message, url string) bool {
	isCrashVideo := checkVideo(url)

	if isCrashVideo {
		if cfg.ReplyToMessage {
			_, _ = session.ChannelMessageSendReply(message.ChannelID, "Crash file detected!", message.Reference())
		}

		err := session.ChannelMessageDelete(message.ChannelID, message.ID)

		if cfg.LogChannel != "" {
			description := ""

			if err != nil {
				description = fmt.Sprintf("**Crash file sent by <@%s> detected in <#%s>**", message.Author.ID, message.ChannelID)
			} else {
				description = fmt.Sprintf("**Crash file sent by <@%s> deleted in <#%s>**", message.Author.ID, message.ChannelID)
			}

			_, _ = session.ChannelMessageSendEmbed(cfg.LogChannel, &discordgo.MessageEmbed{
				Description: description,
				Timestamp:   time.Now().Format(time.RFC3339),
				Color:       11801620,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "AntiCrash :: Author ID " + message.Author.ID,
				},
				Author: &discordgo.MessageEmbedAuthor{
					Name:    message.Author.String(),
					IconURL: message.Author.AvatarURL("128"),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "File URL",
						Value: fmt.Sprintf("`%s`", url),
					},
				},
			})
		}
	}

	return isCrashVideo
}
