package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

type Configuration struct {
	BotToken      string
	ChannelLink   string
	SelectImageID int
	PerHourCount  int
}

func (c *Configuration) setBotToken(botToken string) Configuration {
	c.BotToken = botToken
	viper.Set("BOT_TOKEN", botToken)
	err := viper.WriteConfig()
	if err != nil {
		log.Panicln(err)
	}
	return *c
}

func (c *Configuration) setChannelLink(channelLink string) Configuration {
	c.ChannelLink = channelLink
	viper.Set("CHANNEL_LINK", channelLink)
	err := viper.WriteConfig()
	if err != nil {
		log.Panicln(err)
	}
	return *c
}

func (c *Configuration) setSelectImageID(selectImageID int) Configuration {
	c.SelectImageID = selectImageID
	viper.Set("SELECT_IMAGE_ID", selectImageID)
	err := viper.WriteConfig()
	if err != nil {
		log.Panicln(err)
	}
	return *c
}

func (c *Configuration) setPerHourCount(perHourCount int) Configuration {
	c.PerHourCount = perHourCount
	viper.Set("PER_HOUR_COUNT", perHourCount)
	err := viper.WriteConfig()
	if err != nil {
		log.Panicln(err)
	}
	return *c
}

func loadConfig() Configuration {
	configFileName := "config.toml"
	_, err := os.Stat(configFileName)
	if errors.Is(err, os.ErrNotExist) {
		f, e := os.Create(configFileName)
		if e != nil {
			log.Panicln(e)
		}
		err := f.Close()
		if err != nil {
			log.Panicln(err)
		}
	}

	viper.SetConfigName(configFileName)
	viper.SetConfigType("toml")
	viper.AddConfigPath(filepath.Dir("."))
	err = viper.ReadInConfig()
	if err != nil {
		log.Panicln(err)
	}
	configuration := Configuration{
		viper.GetString("BOT_TOKEN"),
		viper.GetString("CHANNEL_LINK"),
		viper.GetInt("SELECT_IMAGE_ID"),
		viper.GetInt("PER_HOUR_COUNT"),
	}

	if configuration.BotToken == "" {
		fmt.Print("Enter botToken: ")
		var botToken string
		_, err := fmt.Scanln(&botToken)
		if err != nil {
			log.Panicln(err)
		}
		configuration.setBotToken(botToken)
	}

	if configuration.ChannelLink == "" {
		fmt.Print("Enter channelLink(@cat or @thisCatIsBig): ")
		var channelLink string
		_, err := fmt.Scanln(&channelLink)
		if err != nil {
			log.Panicln(err)
		}
		configuration.setChannelLink(channelLink)
	}

	if configuration.ChannelLink == "" {
		fmt.Print("Enter channelLink(@cat or @thisCatIsBig): ")
		var channelLink string
		_, err := fmt.Scanln(&channelLink)
		if err != nil {
			log.Panicln(err)
		}
		configuration.setChannelLink(channelLink)
	}

	if configuration.PerHourCount == 0 {
		fmt.Print("Enter perHourCount(how many pictures will post per hour): ")
		var perHourCount int
		_, err := fmt.Scanln(&perHourCount)
		if err != nil {
			log.Panicln(err)
		}
		configuration.setPerHourCount(perHourCount)
	}

	return configuration
}

func main() {

	configuration := loadConfig()

	bot, err := tgbotapi.NewBotAPI(configuration.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	perHourDelivery(bot, &configuration)

	//updates := bot.GetUpdatesChan(u)
	//for update := range updates {
	//	if update.Message != nil { // If we got a message
	//		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	//
	//		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	//		msg.ReplyToMessageID = update.Message.MessageID
	//
	//		_, err := bot.Send(msg)
	//		if err != nil {
	//			log.Panicln(err)
	//		}
	//	}
	//}
}

type Content struct {
	Text  string
	Image tgbotapi.FilePath
}

func perHourDeliveryGetContentScanFiles(path string) []string {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	var filePathList []string
	for _, f := range files {
		if f.IsDir() {
			filePathList = append(filePathList, perHourDeliveryGetContentScanFiles(path+"/"+f.Name())...)
		} else {
			filePathList = append(filePathList, path+"/"+f.Name())
		}
	}

	return filePathList
}

func perHourDeliveryGetContent(selectImageIdIndex int) Content {

	basePath := "./content"
	filePathList := perHourDeliveryGetContentScanFiles(basePath)

	index := selectImageIdIndex
	file := tgbotapi.FilePath(filePathList[index])
	fileCaption := filePathList[index] + ".caption"
	_, err := os.Stat(fileCaption)
	contentCaption := ""
	if err == nil {
		bytes, err := os.ReadFile(fileCaption)
		if err != nil {
			return Content{}
		}
		contentCaption = string(bytes)
	}
	return Content{contentCaption, file}
}

func perHourDelivery(bot *tgbotapi.BotAPI, configuration *Configuration) {
	// Прошло больше часа с момента отправки последнего сообщения
	var lastMessageAge time.Duration = time.Duration(3600 / configuration.PerHourCount)

	for {
		chatConfig := tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{SuperGroupUsername: configuration.ChannelLink}}
		chat, err := bot.GetChat(chatConfig)
		if err != nil {
			log.Panicln(err)
		}

		if chat.PinnedMessage != nil {
			currentTimeUnix := time.Now().Unix()

			lastMessageTimeUnix := chat.PinnedMessage.Date

			timeMax := int64(lastMessageTimeUnix + 3600/configuration.PerHourCount)

			if timeMax < currentTimeUnix {
				// Act 1. Unpin old messages
				//unpinAllMessagesConfig := tgbotapi.UnpinAllChatMessagesConfig{ChatID: chat.ID}
				//bot.Send(unpinAllMessagesConfig)
				// Act 2. Send new message
				content := perHourDeliveryGetContent(configuration.SelectImageID)
				configuration.setSelectImageID(configuration.SelectImageID + 1)
				contentBaseFile := tgbotapi.BaseFile{BaseChat: tgbotapi.BaseChat{ChatID: chat.ID}, File: content.Image}
				photoPost := tgbotapi.PhotoConfig{BaseFile: contentBaseFile, Caption: content.Text}
				fullSizePhotoPost := tgbotapi.DocumentConfig{BaseFile: contentBaseFile}
				// Фото
				bot.Send(photoPost)
				// И фулл
				bot.Send(fullSizePhotoPost)

				// Act 3. Pin new message
				//pinNewMessageConfig := tgbotapi.PinChatMessageConfig{ChatID: chat.ID, MessageID: photoPostResultMessage.MessageID}
				//bot.Send(pinNewMessageConfig)
				time.Sleep(30 * time.Second)
			}
		} else {
			// Act 2. Send new message
			content := perHourDeliveryGetContent(configuration.SelectImageID)
			configuration.setSelectImageID(configuration.SelectImageID + 1)
			contentBaseFile := tgbotapi.BaseFile{BaseChat: tgbotapi.BaseChat{ChatID: chat.ID}, File: content.Image}
			photoPost := tgbotapi.PhotoConfig{BaseFile: contentBaseFile, Caption: content.Text}
			fullSizePhotoPost := tgbotapi.DocumentConfig{BaseFile: contentBaseFile}
			// Фото
			bot.Send(photoPost)
			// И фулл
			bot.Send(fullSizePhotoPost)
			// Act 3. Pin new message
			//pinNewMessageConfig := tgbotapi.PinChatMessageConfig{ChatID: chat.ID, MessageID: photoPostResultMessage.MessageID}
			//bot.Send(pinNewMessageConfig)
			time.Sleep(lastMessageAge * time.Second)
		}

	}
}
