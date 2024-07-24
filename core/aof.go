package core

import (
	"fmt"
	"log"
	"os"
	"ray/config"
	"strings"
)

func DumpAllAOF() {

	fs, err := os.OpenFile(config.AOFFile, os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("Error opening file : ", config.AOFFile)
		return
	}
	log.Println("re-writing AOF file at : ", config.AOFFile)
	for key, value := range store {
		dumpKey(fs, key, value)
	}

}

func dumpKey(fp *os.File, key string, value *Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, value.Value)
	tokens := strings.Split(cmd, " ")
	fp.Write(Encode(tokens, false))
}
