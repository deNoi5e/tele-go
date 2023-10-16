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
	content, err := ioutil.ReadFile("/mnt/hdd/token") // location token-file
	if err != nil {
		log.Fatal(err)
	}

	var token = string(content)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	const pathForTor = "/mnt/hdd/torrents/" // folder for torrent-files

	log.Printf("Authorized on account %s", bot.Self.UserName)
	_, err = bot.Send(tgbotapi.NewMessage(36327828, "Bot started!"))
	if err != nil {
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			switch update.Message.Text { // Text-commands processing
			case "/ls": // will be removed
				cmd := exec.Command("ls")
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Println(err.Error())
					return
				}
				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, string(stdout)))
				if err != nil {
					log.Panic(err)
					return
				}

			case "/ping": // Test condition bot
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "pong"))
				if err != nil {
					return
				}

			case "/ovpnon": // starting openvpn service on OpenWrt
				cmd := exec.Command("/etc/init.d/openvpn", "start")
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Println(err.Error())
					return
				}

				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "OpenVPN started."+string(stdout)))
				if err != nil {
					return
				}

			case "/ovpnoff": // stopping openvpn service on OpenWrt
				cmd := exec.Command("/etc/init.d/openvpn", "stop")
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Print(err.Error())
					return
				}

				if stdout != nil {

				}

				_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "OpenVPN stopped."))
				if err != nil {
					return
				}

			case "/kill":
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Bot killed."))
				if err != nil {
					return
				}
				cmd := exec.Command("/etc/init.d/telebot", "stop")
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Print(err.Error())
					return
				}

				if stdout != nil {

				}

			case "/reboot": // reboot OpenWrt router
				_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Rebooting..."))
				if err != nil {
					return
				}
				cmd := exec.Command("reboot")
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Print(err.Error())
					return
				}

				if stdout != nil {

				}

			}

			if update.Message.Document != nil { // There is no GetFile method in go-telegram-bot-api. This is its implementation.

				if update.Message.Document.MimeType == "application/x-bittorrent" {
					filename := string(update.Message.Document.FileName)
					fileId := update.Message.Document.FileID

					var link JsonGetFile
					fileLink := "https://api.telegram.org/bot" + token + "/getFile?file_id=" + fileId

					spaceClient := http.Client{
						Timeout: time.Second * 2,
					}

					req, err := http.NewRequest(http.MethodGet, fileLink, nil)
					if err != nil {
						log.Fatal(err)
					}

					req.Header.Set("User-Agent", "spacecount-tutorial")

					res, getErr := spaceClient.Do(req)
					if getErr != nil {
						log.Fatal(getErr)
					}

					if res.Body != nil {
						defer func(Body io.ReadCloser) {
							err := Body.Close()
							if err != nil {

							}
						}(res.Body)
					}

					body, readErr := ioutil.ReadAll(res.Body)
					if readErr != nil {
						log.Fatal(readErr)
					}

					jsonErr := json.Unmarshal(body, &link)
					if jsonErr != nil {
						log.Fatal(jsonErr)
					}

					var downloadLink = "https://api.telegram.org/file/bot" + token + "/" + link.Result.FilePath
					var path = pathForTor + filename
					// var path string = (filename) // debug
					var text = "Torrent file received"

					err = DownloadFile(path, downloadLink)
					if err != nil {
						return
					}
					_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, string(text)))
					if err != nil {
						return
					}
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var out, _ = os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {

		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	return err
}
