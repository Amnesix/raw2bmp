package files

import (
	"io/ioutil"
	"log"
	"strings"
)

func GetRaws(rep string) []string {
	infos, err := ioutil.ReadDir(rep)
	if err != nil {
		log.Fatal("Lecture répertoire !", err)
	}
	var raws []string
	for _, info := range infos {
		if strings.HasSuffix(info.Name(), ".raw") {
			raws = append(raws, info.Name())
		}
	}
	return raws
}
