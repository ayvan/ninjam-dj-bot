package helpers

import (
	"fmt"
	"regexp"
	"strconv"
)

var fileNameNumberedRegexp = regexp.MustCompile(`(.+)(\s+)+(\d+)+\.(\w+){1}$`)
var fileNameRegexp = regexp.MustCompile(`(.+)\.(\w+){1}$`)
var fileMP3Regexp = regexp.MustCompile(`(.+)\.mp3$`)

func IsMP3(fileName string) bool {
	return fileMP3Regexp.MatchString(fileName)
}

func NewFileName(fileName string) (newName string, err error) {
	nameParts := fileNameNumberedRegexp.FindAllStringSubmatch(fileName, -1)

	if len(nameParts) > 0 && len(nameParts[0]) == 5 && nameParts[0][3] != "" && len(nameParts[0][2]) > 0 {

		num, _ := strconv.Atoi(nameParts[0][3])
		num++

		newName = nameParts[0][1] + nameParts[0][2] + strconv.Itoa(num) + "." + nameParts[0][4]
	} else {
		nameParts = fileNameRegexp.FindAllStringSubmatch(fileName, -1)

		if len(nameParts) == 0 || len(nameParts[0]) != 3 {
			err = fmt.Errorf("bad file name")
			return
		}

		newName = nameParts[0][1] + " 2." + nameParts[0][2]
	}

	return
}
