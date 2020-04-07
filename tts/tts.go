// Copied from github.com/guumaster/go-tts/
package tts

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gosimple/slug"
	"golang.org/x/text/language"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const baseDir = "/tmp"

func googleTTSReader(text, lang string, slow bool) (io.ReadCloser, error) {
	speed := "1"
	if slow {
		speed = "0.24"
	}

	q := url.Values{}
	q.Set("ie", "UTF-8")
	q.Set("total", "1")
	q.Set("idx", "0")
	q.Set("client", "tw-ob")
	q.Set("tl", lang)
	q.Set("ttsspeed", speed)
	q.Set("q", text)
	q.Set("textlen", strconv.Itoa(len(text)))

	u := &url.URL{
		Scheme:   "https",
		Host:     "translate.google.com",
		Path:     "translate_tts",
		RawQuery: q.Encode(),
	}

	response, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

func Say(l, text string, slow bool) ([]byte, error) {
	ltag, err := language.Parse(l)
	if err != nil {
		return nil, err
	}
	lang := ltag.String()

	key := key(fmt.Sprintf("%s_%s_%t", text, lang, slow), lang)
	path := path(fmt.Sprintf("%s.mp3", key))

	var audio io.ReadCloser
	if exist(path) {
		audio, err := open(path)
		if err != nil {
			return nil, err
		}
		defer audio.Close()
	} else {
		audio, err = googleTTSReader(text, lang, slow)
		if err != nil {
			return nil, err
		}
		defer audio.Close()
	}

	var audioBuf bytes.Buffer
	audioReader := io.TeeReader(audio, &audioBuf)

	b, err := ioutil.ReadAll(audioReader)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("audio buffer is empty")
	}

	err = save(path, audioBuf)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func key(in, lang string) string {
	h := md5.New()
	h.Write([]byte(slug.MakeLang(in, lang)))
	return hex.EncodeToString(h.Sum(nil))
}

func path(filename string) string {
	return fmt.Sprintf("%s/%s", baseDir, filename)
}

func open(filename string) (io.ReadCloser, error) {
	path := path(filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func save(path string, content bytes.Buffer) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(content.Bytes())

	return err
}

func exist(filename string) bool {
	path := path(filename)
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
