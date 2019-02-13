package main

import (
	"fmt"
	"github.com/Ayvan/ninjam-chatbot/models"
	"github.com/Ayvan/ninjam-chatbot/ninjam-bot"
	"github.com/Ayvan/ninjam-dj-bot/api"
	"github.com/Ayvan/ninjam-dj-bot/config"
	"github.com/Ayvan/ninjam-dj-bot/dj"
	"github.com/Ayvan/ninjam-dj-bot/tracks"
	"github.com/Ayvan/ninjam-dj-bot/tracks_sync"
	"github.com/VividCortex/godaemon"
	"github.com/burillo-se/lv2hostconfig"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	config.Init()

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

	bot.OnSuccessAuth(func() {
		bot.ChannelInit("BackingTrack")
	})

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
		bot.Connect()
	}(bot)

	go func(bot *ninjam_bot.NinJamBot) {
		for {
			select {
			case msg := <-bot.IncomingMessages():
				bm := BotMessage{
					Bot:     bot,
					Message: msg,
				}
				botChan <- bm
			case <-sigChan:
				sigChan <- true
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
			if strings.HasPrefix(msg.Message.Name, msg.Bot.UserName()) {
				continue
			}
			logrus.Info(fmt.Sprintf("%s: %s", msg.Message.Name, msg.Message.Text))

			switch msg.Message.Type {
			case models.PART:
				timer := time.NewTimer(time.Second * 5)
				go func() {
					<-timer.C
					logrus.Info("Users after part: ", len(bot.Users()), bot.Users())
					if len(bot.Users()) == 1 {
						logrus.Info("Stop player: only 1 user on server, it must be jamtrack bot")
						// TODO dj.StopMP3()
					}
				}()
			case models.MSG:
				msg := jamManager.Command(msg.Message.Text)
				bot.SendMessage(msg)
			}
		}
	}

	wg.Wait()

	// t.SendMessage(fmt.Sprint("Новые участники на джем-сервере ", n.Name, ": ", liUsers))
	// t.SendMessage(fmt.Sprint("Джем-сервер ", n.Name, " покинули: ", loUsers))

	logrus.Info("Application ", config.Get().AppName, " finished")
}
