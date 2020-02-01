package dj

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQueueManager_Add(t *testing.T) {
	qm := NewQueueManager("dj", func(string) {})

	assert.Equal(t, uint(0), qm.UsersCount())

	qm.Add("test1")
	assert.NotNil(t, qm.current)
	assert.Nil(t, qm.current.Next)
	assert.Equal(t, uint(1), qm.UsersCount())

	qm.Add("test2")
	assert.NotNil(t, qm.current.Next)
	assert.Equal(t, uint(2), qm.UsersCount())

	qm.Add("test3")
	assert.NotNil(t, qm.current.Next.Next)
	assert.Equal(t, uint(3), qm.UsersCount())

	qm.Add("test4")
	assert.Equal(t, uint(4), qm.UsersCount())

	qm.Add("test3")
	assert.Equal(t, uint(4), qm.UsersCount())

	qm.Add("test4")
	assert.Equal(t, uint(4), qm.UsersCount())

	qm.Add("test1")
	assert.Equal(t, uint(4), qm.UsersCount())
	assert.Equal(t, "test2", qm.current.Name)

	qm.Add("dj")
	assert.Equal(t, uint(4), qm.UsersCount())

	qm.Add("dj@127.x.x.1")
	assert.Equal(t, uint(4), qm.UsersCount())
}

func TestQueueManager_queue(t *testing.T) {
	qm := NewQueueManager("dj", func(string) {})

	qm.Add("burillo")
	assert.Nil(t, qm.current.Prev)
	qm.Add("cronos")
	assert.Nil(t, qm.current.Prev)
	qm.Add("archi")
	assert.Nil(t, qm.current.Prev)

	qm.Del("burillo")
	assert.Nil(t, qm.current.Prev)

	qm.Add("burillo")
	assert.Nil(t, qm.current.Prev)

	qm.next()
	assert.Nil(t, qm.current.Prev)

	qm.next()
	assert.Nil(t, qm.current.Prev)

	qm.next()
	assert.Nil(t, qm.current.Prev)

	qm.Del("cronos")
	assert.Nil(t, qm.current.Prev)
	assert.Nil(t, qm.current.Next.Next)
	qm.Del("archi")
	assert.Nil(t, qm.current.Prev)
}

func TestQueueManager_Del(t *testing.T) {
	qm := NewQueueManager("dj", func(string) {})

	assert.Equal(t, uint(0), qm.UsersCount())

	qm.Add("test1")
	assert.Equal(t, uint(1), qm.UsersCount())

	qm.Add("test2")
	assert.Equal(t, uint(2), qm.UsersCount())

	qm.Add("test3")
	assert.Equal(t, uint(3), qm.UsersCount())

	qm.Add("test4")
	assert.Equal(t, uint(4), qm.UsersCount())

	qm.Del("test2")
	assert.Equal(t, uint(3), qm.UsersCount())

	qm.Del("test4")
	assert.Equal(t, uint(2), qm.UsersCount())

	qm.Del("test2")
	assert.Equal(t, uint(2), qm.UsersCount())

	qm.Del("test4")
	assert.Equal(t, uint(2), qm.UsersCount())

	qm.Add("test2")
	assert.Equal(t, uint(3), qm.UsersCount())

	qm.Add("dj@101.x.x.203")
	assert.Equal(t, uint(3), qm.UsersCount())

	qm.Add("dj")
	assert.Equal(t, uint(3), qm.UsersCount())

	assert.Equal(t, "test1", qm.current.Name)
	assert.Equal(t, "test3", qm.current.Next.Name)
	qm.Del("test1")
	assert.Equal(t, uint(2), qm.UsersCount())
	assert.Equal(t, "test3", qm.current.Name)
	assert.Equal(t, "test2", qm.current.Next.Name)

	qm.Del("test2")
	assert.Equal(t, uint(1), qm.UsersCount())
	assert.Equal(t, "test3", qm.current.Name)
	assert.Nil(t, qm.current.Prev)
	assert.Nil(t, qm.current.Next)

	qm.Del("test3")
	assert.Equal(t, uint(0), qm.UsersCount())
	assert.Nil(t, qm.current)
}

func TestQueueManager_Del_2(t *testing.T) {
	qm := NewQueueManager("dj", func(string) {})

	assert.Equal(t, uint(0), qm.UsersCount())

	qm.Add("test1")
	assert.Equal(t, uint(1), qm.UsersCount())

	qm.Add("test2")
	assert.Equal(t, uint(2), qm.UsersCount())

	qm.Add("test3")
	assert.Equal(t, uint(3), qm.UsersCount())

	qm.Add("test4")
	assert.Equal(t, uint(4), qm.UsersCount())

	qm.Del("test1")
	assert.Equal(t, uint(3), qm.UsersCount())

	assert.Equal(t, "test2", qm.current.Name)
	assert.Equal(t, "test3", qm.current.Next.Name)

	qm.Add("test1")
	assert.Equal(t, uint(4), qm.UsersCount())

	qm.Del("test1")
	assert.Equal(t, uint(3), qm.UsersCount())

	qm.Add("test1")
	assert.Equal(t, uint(4), qm.UsersCount())

	qm.Del("test2")
	assert.Equal(t, uint(3), qm.UsersCount())

	qm.Del("test1")
	assert.Equal(t, uint(2), qm.UsersCount())

	qm.Add("test1")
	assert.Equal(t, uint(3), qm.UsersCount())
}

func TestNewQueueManager(t *testing.T) {
	t.Skip()
	qm := NewQueueManager("dj", func(msg string) {
		fmt.Println("==", msg)
	})

	qm.Add("User 1")
	qm.Add("User 2")
	qm.Add("User 3")

	qm.OnStart(time.Minute*20, time.Second*10)
	qm.userPlayDuration = time.Second * 30

	time.Sleep(time.Second * 10)
	qm.Del("User 1")
	qm.Del("User 3")

	time.Sleep(time.Second)
	qm.Add("User 1")
	qm.Add("User 3")

	time.Sleep(time.Hour)
}

func Test_cleanName(t *testing.T) {
	assert.Equal(t, "dj", cleanName("dj@127.x.x.1"))
	assert.Equal(t, "dj", cleanName("dj@210.x.x.101"))
	assert.Equal(t, "dj", cleanName("dj"))
}
