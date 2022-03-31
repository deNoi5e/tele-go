package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"fmt"
	"os/exec"
)

type JsonGetFile struct {
	Ok     bool `json:"ok"`
	Result struct {
		FileID       string `json:"file_id"`
		FileUniqueID string `json:"file_unique_id"`
		FileSize     int    `json:"file_size"`
		FilePath     string `json:"file_path"`
	} `json:"result"`
}

func main() {
	content, err := ioutil.ReadFile("token")
	if err != nil {
		log.Fatal(err)
	}

	tgtoken := string(content)

	bot, err := tgbotapi.NewBotAPI(tgtoken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	const path_for_tor = "/mnt/hdd/torrents/"

	log.Printf("Authorized on account %s", bot.Self.UserName)
	bot.Send(tgbotapi.NewMessage(36327828, "Bot started!"))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			switch update.Message.Text { // Text-commands processing
			case "/ls":
				cmd := exec.Command("ls")
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Println(err.Error())
					return
				}
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, string(stdout)))

			case "/ping":
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "pong"))

			case "/ovpnon":
				cmd := exec.Command("/etc/init.d/openvpn", "start")
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Println(err.Error())
					return
				}

				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "OpenVPN started."+string(string(stdout))))

			case "/ovpnoff":
				cmd := exec.Command("/etc/init.d/openvpn", "stop")
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Print(err.Error())
					return
				}

				if stdout != nil {

				}

				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "OpenVPN stopped."))

			case "/reboot":
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Rebooting..."))
				cmd := exec.Command("reboot")
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Print(err.Error())
					return
				}

				if stdout != nil {

				}

			}

			if update.Message.Document != nil {

				if update.Message.Document.MimeType == "application/x-bittorrent" {
					filename := string(update.Message.Document.FileName)
					fileid := update.Message.Document.FileID

					var link JsonGetFile
					var filelink string = ("https://api.telegram.org/bot" + tgtoken + "/getFile?file_id=" + fileid)

					spaceClient := http.Client{
						Timeout: time.Second * 2,
					}

					req, err := http.NewRequest(http.MethodGet, filelink, nil)
					if err != nil {
						log.Fatal(err)
					}

					req.Header.Set("User-Agent", "spacecount-tutorial")

					res, getErr := spaceClient.Do(req)
					if getErr != nil {
						log.Fatal(getErr)
					}

					if res.Body != nil {
						defer res.Body.Close()
					}

					body, readErr := ioutil.ReadAll(res.Body)
					if readErr != nil {
						log.Fatal(readErr)
					}

					jsonErr := json.Unmarshal(body, &link)
					if jsonErr != nil {
						log.Fatal(jsonErr)
					}

					var download_link string = ("https://api.telegram.org/file/bot" + tgtoken + "/" + link.Result.FilePath)
					var path string = (path_for_tor + filename)
					// var path string = (filename) // debug
					var text string = "Torrent file received"

					DownloadFile(path, download_link)
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, string(text)))
				}

			}
		}
	}
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
