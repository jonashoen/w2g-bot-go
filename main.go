package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type VideoSearchResult struct {
    Url string
    Title string
    Thumb string
}

type CreateRoomRequest struct {
    Share string `json:"share"`
    Title string `json:"title"`
    Thumb string `json:"thumb"`
}

type Room struct {
    Streamkey string
}

func main() {
	token, foundToken := os.LookupEnv("BOT_TOKEN")

	if !foundToken {
		fmt.Println("Discord API token not found!")
        return
	}

    // Create a new Discord session using the provided bot token.
    discord, createDiscordBotError := discordgo.New("Bot " + token)
    if createDiscordBotError != nil {
        fmt.Println("Error creating Discord session,", createDiscordBotError)
        return
    }

    // Register the messageCreate func as a callback for MessageCreate events.
    discord.AddHandler(messageCreate)

    // In this example, we only care about receiving message events.
    discord.Identify.Intents = discordgo.IntentsGuildMessages

    // Open a websocket connection to Discord and begin listening.
    startDiscordBotError := discord.Open()
    if startDiscordBotError != nil {
        fmt.Println("Error opening connection,", startDiscordBotError)
        return
    }

    // Wait here until CTRL-C or other term signal is received.
    fmt.Println("Bot is now running.")

    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
    <-sc

    // Cleanly close down the Discord session.
    discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Ignore all messages created by the bot itself
    // This isn't required in this specific example but it's a good practice.
    if m.Author.ID == s.State.User.ID {
        return
    }

    // every message needs to be prefixed with !w2g
    if strings.HasPrefix(m.Content, "!w2g") {
		videoUrl := strings.TrimSpace(
            strings.ReplaceAll(m.Content, "!w2g", ""),
        )
        
        if(videoUrl == "") {
            // empty links create an empty room
            response, err := http.Post("https://w2g.tv/rooms/create", "application/json", nil)

            if err != nil {
                fmt.Println(err)
            }

            defer response.Body.Close()
        
            // send message with room link
            s.ChannelMessageSend(
                m.ChannelID,
                response.Request.URL.String(),
            )
        } else {
            // if a link is provided the video will be autoplayed in the room
            if !strings.Contains(videoUrl, "https") {
                videoUrl = "https://www.youtube.com/watch?v=" + videoUrl
            }

            videoInfo := searchVideo(videoUrl)
            roomInfo := CreateRoom(videoInfo)

            // send message with room link
            s.ChannelMessageSend(
                m.ChannelID,
                "https://w2g.tv/rooms/" + roomInfo.Streamkey,
            )
        }
    }
}

func searchVideo(videoUrl string) VideoSearchResult {
    // look up video by url
    response, lookUpError := http.Get("https://w2g.tv/w2g_search/lookup?url=" + videoUrl)

    if lookUpError != nil {
        fmt.Println(lookUpError)
    }

    // read the response into the body buffer
    body, readResponseError := io.ReadAll(response.Body)

    defer response.Body.Close()

    if readResponseError != nil {
        fmt.Println(readResponseError)
    }

    var videoInfo VideoSearchResult

    // read the response into the videoInfo struct
    videoInfoParsingError := json.Unmarshal(body, &videoInfo)

    if videoInfoParsingError != nil {
        fmt.Println(videoInfoParsingError)
    }

    return videoInfo
}

func CreateRoom(videoInfo VideoSearchResult) Room {
    // create struct for room creation
    createRoomRequest := CreateRoomRequest{Share: videoInfo.Url, Title: videoInfo.Title, Thumb: videoInfo.Thumb}

    // parse struct to byte array
    rawCreateRoomRequest, createRoomError := json.Marshal(createRoomRequest)

    if createRoomError != nil {
        fmt.Println(createRoomError)
    }

    // send the create room request
    createRoomResponse, lookUpError := http.Post(
        "https://w2g.tv/rooms/create.json",
        "application/json",
        bytes.NewReader(rawCreateRoomRequest),
    )

    if lookUpError != nil {
        fmt.Println(lookUpError)
    }

    // read the response into the roomBody buffer
    roomBody, readResponseError := io.ReadAll(createRoomResponse.Body)

    defer createRoomResponse.Body.Close()

    if readResponseError != nil {
        fmt.Println(readResponseError)
    }

    var room Room

    // parse the response into the room struct
    roomInfoParsingError := json.Unmarshal(roomBody, &room)

    if roomInfoParsingError != nil {
        fmt.Println(roomInfoParsingError)
    }

    return room
}