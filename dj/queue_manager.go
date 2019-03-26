package dj

import (
	"github.com/ayvan/ninjam-chatbot/models"
	"github.com/ayvan/ninjam-dj-bot/lib"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"strings"
	"sync"
	"time"
)

const (
	messageAfter15Seconds = "%s's turn in 15 seconds"
	messageNowPlaying     = "%s is playing now"
	messageIsNext         = "%s is next"
)

func init() {
	message.SetString(language.Russian, messageAfter15Seconds, "очередь %s через 15 секунд")
	message.SetString(language.Russian, messageNowPlaying, "сейчас играет %s")
	message.SetString(language.Russian, messageIsNext, "готовится играть %s")
}

type QueueManager struct {
	botName           string
	userStartTime     *time.Time
	userStartsPlaying *user

	userPlayDuration  time.Duration
	trackEndTime      time.Time
	sendMessage       func(msg string)
	first             *user
	current           *user
	after15SecMsgSent bool // флаг что сообщение messageAfter15Seconds уже отправлено
	mtx               *sync.Mutex

	stopped     bool
	stopChannel chan bool
}

type user struct {
	Name string
	Prev *user
	Next *user
}

func NewQueueManager(botName string, sendMessageFunc func(msg string)) *QueueManager {
	qm := &QueueManager{botName: botName, sendMessage: sendMessageFunc}
	qm.stopChannel = make(chan bool, 1)
	qm.mtx = new(sync.Mutex)
	go qm.supervisor()

	return qm
}

func (qm *QueueManager) Close() {
	qm.stopChannel <- true
	close(qm.stopChannel)
}

func (qm *QueueManager) supervisor() {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			if qm.stopped {
				continue
			}
			if qm.userStartTime == nil {
				continue
			}
			// если до конца трека осталось менее чем qm.userPlayDuration то ничего не делаем
			if qm.trackEndTime.After(time.Now()) && qm.trackEndTime.Sub(time.Now()) < time.Second*15 {
				continue
			}
			if qm.userStartTime.Add(qm.userPlayDuration).Before(time.Now()) &&
				qm.userStartTime.Add(qm.userPlayDuration + time.Second*15).After(time.Now()) {
				if qm.current != nil && qm.current.Next != nil && qm.sendMessage != nil && !qm.after15SecMsgSent {
					qm.sendMessage(p.Sprintf(messageAfter15Seconds, qm.current.Next.Name))
					qm.after15SecMsgSent = true
				}
				continue
			}
			if qm.userStartTime.Add(qm.userPlayDuration).After(time.Now()) {
				continue
			}
			qm.next()
		case <-qm.stopChannel:
			ticker.Stop()
			return
		}
	}
}

func (qm *QueueManager) UsersCount() (i uint) {
	users := qm.Users()
	return uint(len(users))
}

func (qm *QueueManager) Users() (users []string) {
	if qm.current == nil {
		return
	}

	i := 0
	curr := qm.current
	for {
		users = append(users, curr.Name)
		i++
		if curr.Next == nil {
			return
		}
		curr = curr.Next

		// shit happened...
		if i > 1000 {
			i = 0
			return
		}
	}
}

func (qm *QueueManager) Add(userName string) {
	userName = cleanName(userName)
	if userName == qm.botName {
		return
	}

	qm.Del(userName)

	qm.mtx.Lock()
	defer qm.mtx.Unlock()
	logrus.Debugf("user %s joined", userName)

	newUser := &user{Name: userName}
	if qm.current == nil {
		qm.current = newUser
		qm.first = newUser
		return
	}

	i := 0
	curr := qm.current
	for {
		if curr.Next == nil {
			curr.Next = newUser
			curr.Next.Prev = curr
			return
		}
		curr = curr.Next

		if curr == qm.current {
			logrus.Error("Shit happened! QueueManager.Add curr == qm.current")
			return
		}

		i++
		if i > 1000 {
			logrus.Error("Shit happened! QueueManager.Add")
			return
		}
	}
}

func (qm *QueueManager) Del(userName string) {
	qm.mtx.Lock()
	defer qm.mtx.Unlock()
	logrus.Debugf("user %s leaved", userName)
	defer func() {
		if qm.current != nil {
			logrus.Debugf("current user %s", qm.current.Name)

			if qm.current.Next != nil {
				logrus.Debugf("next user %s", qm.current.Next.Name)
			}
		}
	}()

	userName = cleanName(userName)
	if userName == qm.botName {
		return
	}
	if qm.current == nil {
		return
	}
	curr := qm.current
	i := 0
	for {
		if curr == nil {
			return
		}
		if curr.Name == userName {
			if curr.Prev != nil {
				curr.Prev.Next = curr.Next
			}
			if curr.Next == nil && i == 0 {
				qm.current = nil
				return
			}

			if curr.Next != nil {
				curr.Next.Prev = curr.Prev
			}

			if i == 0 {
				if curr.Next == nil {
					// TODO остановить плеер
				}
				qm.current = curr.Next
				qm.current.Prev = nil
				// если текущий юзер и есть выбывший - сразу переключаем
				qm.userStartTime = nil
				qm.userStartsPlaying = nil
				qm.start(0)
			}

			return
		}
		curr = curr.Next

		if curr == qm.current {
			logrus.Error("Shit happened! QueueManager.Del curr == qm.current")
			return
		}

		i++
		if i > 1000 {
			logrus.Error("Shit happened! QueueManager.Del i > 1000")
			return
		}
	}
}

func (qm *QueueManager) next() {
	qm.mtx.Lock()
	defer qm.mtx.Unlock()

	if qm.current != nil && qm.current.Next != nil {
		next := qm.current.Next

		curr := qm.current
		// перекинем текущего в конец списка
		i := 0
		for {
			if curr.Next != nil {
				if curr == curr.Next {
					logrus.Error("Shit happened! QueueManager.next current == current.Next")
					curr.Next = nil
					break
				}

				curr = curr.Next

				if curr == qm.current {
					logrus.Error("Shit happened! QueueManager.next curr == qm.current")
					return
				}

				i++
				if i > 1000 {
					logrus.Error("Shit happened! QueueManager.next i > 1000")
					return
				}
				continue
			}
			curr.Next = qm.current
			curr.Next.Prev = curr
			curr.Next.Next = nil
			break
		}
		qm.current = next
		qm.userStartTime = nil
		qm.userStartsPlaying = nil
		qm.start(0)
		return
	}

	// если следующего нет - просто обновим таймер и текущий продолжит играть
	tn := time.Now()
	qm.userStartTime = &tn
}

func (qm *QueueManager) start(intervalDuration time.Duration) {
	//  если уже кто-то играл - переключим на следующего на новом треке
	if qm.userStartTime != nil &&
		qm.current != nil &&
		qm.userStartsPlaying == qm.current && // may be different if current user leaved server and next user has become current
		qm.userStartTime.Add(qm.userPlayDuration).Before(time.Now()) &&
		qm.current.Next != nil {
		qm.userStartTime = nil
		qm.userStartsPlaying = nil
		qm.next()
		return
	}
	tn := time.Now().Add(intervalDuration)
	qm.userStartTime = &tn
	qm.userStartsPlaying = qm.current
	qm.after15SecMsgSent = false
	qm.stopped = false
	if qm.current != nil && qm.sendMessage != nil {
		// если до конца трека осталось примерно время игры одного музыканта - не объявляем следующего
		if qm.current.Next == nil || time.Now().Add(qm.userPlayDuration + time.Second*10).After(qm.trackEndTime) {
			qm.sendMessage(p.Sprintf(messageNowPlaying, qm.current.Name))
		} else {
			qm.sendMessage(p.Sprintf(messageNowPlaying, qm.current.Name) + ", " + p.Sprintf(messageIsNext, qm.current.Next.Name))
		}
	}
}

func (qm *QueueManager) OnStart(trackDuration, intervalDuration time.Duration) {
	qm.userPlayDuration = lib.CalcUserPlayDuration(trackDuration)
	qm.trackEndTime = time.Now().Add(trackDuration)
	qm.start(intervalDuration)
}

func (qm *QueueManager) OnStop() {
	qm.stopped = true
}

func (qm *QueueManager) OnUserinfoChange(user models.UserInfo) {
	if user.Active == 0x1 {
		qm.Add(string(user.Name))
		return
	}
	qm.Del(string(user.Name))
}

func cleanName(userName string) string {
	i := strings.Index(userName, "@")
	if i < 0 {
		return userName
	}

	return userName[:i]
}
