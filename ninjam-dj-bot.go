package main

import (
	"fmt"
	"github.com/VividCortex/godaemon"
	"github.com/ayvan/ninjam-chatbot/models"
	"github.com/ayvan/ninjam-chatbot/ninjam-bot"
	"github.com/ayvan/ninjam-dj-bot/api"
	"github.com/ayvan/ninjam-dj-bot/config"
	"github.com/ayvan/ninjam-dj-bot/dj"
	"github.com/ayvan/ninjam-dj-bot/tracks"
	"github.com/ayvan/ninjam-dj-bot/tracks_sync"
	"github.com/burillo-se/lv2hostconfig"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	if config.Get().DaemonMode {
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	}

	go api.Run("0.0.0.0:" + config.Get().HTTPPort)

	jamDB, err := tracks.NewJamDB(config.Get().DBFile)
	if err != nil {
		logrus.Fatal(err)
	}
	api.Init(jamDB)

	pidFile := config.Get().AppPidPath

	if pidFile != "" {

		pid := fmt.Sprintf("%d", os.Getpid())

		err := ioutil.WriteFile(pidFile, []byte(pid), 0644)

		if err != nil {
			logrus.Fatal("Error when writing pidfile:", err)
		}

		defer func() {
			os.Remove(pidFile)
		}()
	}

	lv2hostConfigPath := config.Get().LV2HostConfig

	hostconfig := lv2hostconfig.NewLV2HostConfig()
	hostconfig.ReadFile(lv2hostConfigPath)

	sChan := make(chan os.Signal, 1)
	// ловим команды на завершение от ОС и корректно завершаем приложение с помощью sync.WaitGroup
	signal.Notify(sChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	server := config.Get().Server

	bot := ninjam_bot.NewNinJamBot(server.Host, server.Port, server.UserName, server.UserPassword, server.Anonymous)

	dir, err := filepath.Abs(config.Get().TracksDir)
	if err != nil {
		logrus.Fatal(err)
	}

	jp := dj.NewJamPlayer(dir, bot, hostconfig)

	tracks_sync.Init(dir, jamDB)

	bot.SetOnSuccessAuth(func() {
		bot.ChannelInit("BackingTrack")
	})

	bot.SetOnServerConfigChange(jp.OnServerConfigChange)

	jamManager := dj.NewJamManager(jamDB, jp, bot)

	// инициализируем глобальный канал завершения горутин
	sigChan := make(chan bool, 1)

	go func() {
		// ловим сигнал завершения, выводим информацию в лог, а затем отправляем его в глобальный канал
		s := <-sChan
		logrus.Info("os.Signal ", s, " received, finishing application...")
		bot.Stop()
		sigChan <- true
		return
	}()

	wg := &sync.WaitGroup{}

	logrus.Info("Application ", config.Get().AppName, " started")

	type BotMessage struct {
		Bot     *ninjam_bot.NinJamBot
		Message models.Message
	}

	botChan := make(chan BotMessage, 1000)

	wg.Add(1)
	go func(bot *ninjam_bot.NinJamBot) {
		defer wg.Done()
		defer logrus.Info("bot.Connect() finished")
		bot.Connect()
	}(bot)

	go func(bot *ninjam_bot.NinJamBot) {
		defer logrus.Info("DJ bot IncomingMessages loop finished")
		for {
			select {
			case msg := <-bot.IncomingMessages():
				bm := BotMessage{
					Bot:     bot,
					Message: msg,
				}
				logrus.Debug("IncomingMessage %s", msg)
				botChan <- bm
				logrus.Debug("IncomingMessage sent to botChan %s", msg)
			case <-sigChan:
				sigChan <- true
				defer logrus.Info("DJ bot IncomingMessages loop sigChan received")
				return
			}
		}
	}(bot)

f:
	for {
		select {
		case s := <-sigChan:
			sigChan <- s
			break f
			// messages routers <->
		case msg := <-botChan:
			logrus.Debug("botChan received message %s", msg)
			if strings.HasPrefix(msg.Message.Name, msg.Bot.UserName()) || msg.Message.Name == "" {
				continue
			}
			logrus.Info(fmt.Sprintf("%s: %s", msg.Message.Name, msg.Message.Text))

			switch msg.Message.Type {
			case models.PART:
				timer := time.NewTimer(time.Second * 5)
				go func() {
					<-timer.C
					logrus.Info("Users after part: ", len(bot.Users()), bot.Users())
				}()
			case models.MSG:
				r := regexp.MustCompile(`^` + bot.UserName() + `\s+(.*)`)
				s := r.FindStringSubmatch(msg.Message.Text)

				command := ""

				if len(s) > 0 {
					command = s[1]
					if command != "" {
						msg := jamManager.Command(command)
						bot.SendMessage(msg)
					}
				}
			}
		}
	}
	logrus.Info("DJ bot main loop finished")
	wg.Wait()

	// t.SendMessage(fmt.Sprint("Новые участники на джем-сервере ", n.Name, ": ", liUsers))
	// t.SendMessage(fmt.Sprint("Джем-сервер ", n.Name, " покинули: ", loUsers))

	logrus.Info("Application ", config.Get().AppName, " finished")
}
