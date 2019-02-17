package dj

import (
	"github.com/ayvan/ninjam-chatbot/models"
	"time"
)

type QueueManager struct {
	userStartTime    *time.Time
	userPlayDuration time.Duration
	sendMessage      func(msg string)
	first            *user
	current          *user
	stopped          bool
	stopChannel      chan bool
}

type user struct {
	Name string
	Prev *user
	Next *user
}

func NewQueueManager(sendMessageFunc func(msg string)) *QueueManager {
	qm := &QueueManager{sendMessage: sendMessageFunc}
	qm.stopChannel = make(chan bool, 1)
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
			if qm.userStartTime.Add(qm.userPlayDuration).After(time.Now()) && qm.userStartTime.Add(qm.userPlayDuration + time.Second*15).Before(time.Now()) {
				// TODO announce следующий через 15 секунд
				continue
			}
			if qm.userStartTime.Add(qm.userPlayDuration).After(time.Now()) {
				continue
			}
			// TODO если до конца трека осталось менее чем qm.userPlayDuration то ничего не делаем
			qm.next()
		case <-qm.stopChannel:
			ticker.Stop()
			return
		}
	}
}

func (qm *QueueManager) Add(userName string) {
	newUser := &user{Name: userName}
	if qm.current == nil {
		qm.current = newUser
		qm.first = newUser
		return
	}

	curr := qm.current
	for {
		if curr.Next == nil {
			curr.Next = newUser
			curr.Next.Prev = curr
			return
		}
		curr = qm.current.Next
	}
}

func (qm *QueueManager) Del(userName string) {
	if qm.current == nil {
		return
	}
	curr := qm.current
	i := 0
	for {
		if curr.Name == userName {
			curr.Prev.Next = curr.Next
			curr.Next.Prev = curr.Prev
			qm.current = curr.Next

			if i == 0 {
				// если текущий юзер и есть выбывший - сразу переключаем
				qm.next()
			}
			return
		}
		curr = qm.current.Next
		i++
	}
}

func (qm *QueueManager) next() {
	if qm.current.Next != nil {
		qm.current = qm.current.Next
		qm.start(0)
		return
	}

	// если следующего нет - просто обновим таймер и текущий продолжит играть
	tn := time.Now()
	qm.userStartTime = &tn
}

func (qm *QueueManager) start(intervalDuration time.Duration) {
	tn := time.Now().Add(intervalDuration)
	qm.userStartTime = &tn
	// TODO announce
}

func (qm *QueueManager) OnStart(trackDuration, intervalDuration time.Duration) {
	// TODO по длине трека рассчитывать userPlayDuration и сохранять инфу о времени когда трек закончится
	// TODO дёргать из плеера когда запустили стрим интервала
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
