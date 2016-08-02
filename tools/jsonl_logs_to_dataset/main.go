package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

type Message struct {
	Event   string `json:"event"`
	ID      string `json:"id"`
	Service bool   `json:"service"`
	Flags   int    `json:"flags"`
	ReplyID string `json:"reply_id"`
	Unread  bool   `json:"unread"`
	To      struct {
		PeerType   string `json:"peer_type"`
		ID         string `json:"id"`
		MembersNum int    `json:"members_num"`
		PeerID     int    `json:"peer_id"`
		PrintName  string `json:"print_name"`
		Flags      int    `json:"flags"`
		Title      string `json:"title"`
		Admin      struct {
			PeerType  string `json:"peer_type"`
			ID        string `json:"id"`
			PeerID    int    `json:"peer_id"`
			PrintName string `json:"print_name"`
		} `json:"admin"`
	} `json:"to"`
	From struct {
		When      string `json:"when"`
		Username  string `json:"username"`
		ID        string `json:"id"`
		PeerType  string `json:"peer_type"`
		FirstName string `json:"first_name"`
		PeerID    int    `json:"peer_id"`
		PrintName string `json:"print_name"`
		Flags     int    `json:"flags"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	} `json:"from"`
	Text  string `json:"text"`
	Out   bool   `json:"out"`
	Date  int    `json:"date"`
	Media struct {
		Type    string `json:"type"`
		Caption string `json:"caption"`
	} `json:"media"`
}

type ReadLogRes struct {
	Error    error
	Success  bool
	Output   []byte
	FileName string
}

var inFile = flag.String("in", "./log.jsonl", "the filepaths in a comma seperated list to the jsonl telegram logs")
var outFile = flag.String("out", "./dataset.txt", "the dataset output")

func main() {
	flag.Parse()

	logFiles := strings.Split(*inFile, ",")

	outChan := make(chan ReadLogRes, len(logFiles))

	for _, logFile := range logFiles {
		go readLogFile(outChan, logFile)
	}

	f, err := os.Create(*outFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var numDone int
	for {
		res := <-outChan

		if res.Success {
			log.Println("Writing", res.FileName, "to out")
			_, err := f.Write(res.Output)
			if err != nil {
				log.Println("Failed to write", res.FileName, "to out")
			} else {
				log.Println("Wrote", res.FileName, "to out")
			}

		} else {
			log.Println("Result was not a success:", res.Error)
		}

		numDone++
		if numDone >= len(logFiles) {
			break
		}
	}

	log.Println("Finished Converting Logs. Quitting...")
}

func readLogFile(outChan chan ReadLogRes, fn string) {
	var err error
	var res []byte

	defer func() {
		outChan <- ReadLogRes{
			Error:    err,
			Success:  err == nil && len(res) > 0,
			Output:   res,
			FileName: fn,
		}
	}()

	fileBytes, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Println("Error! Could not read file:", fn, "With error:", err)
		return
	}

	msgs := bytes.Split(fileBytes, []byte{'\n'})

	msgCount := 0
	for _, msgBytes := range msgs {
		if len(msgBytes) > 3 {
			msgCount++
		}
	}

	msgTexts := make([]string, msgCount)

	wait := new(sync.WaitGroup)
	for i, msgBytes := range msgs {
		wait.Add(1)

		go func(j int, msgB []byte, fileName string) {
			defer wait.Done()
			msgTxt := doMsgDecode(msgB, fileName)
			if msgTxt != "" {
				msgTexts[j] = msgTxt
			}
		}(i, msgBytes, fn)
	}

	for i := len(msgTexts)/2-1; i >= 0; i-- {
		opp := len(msgTexts)-1-i
		msgTexts[i], msgTexts[opp] = msgTexts[opp], msgTexts[i]
	}

	res = []byte(strings.Join(msgTexts, "\n"))
	log.Println("Finished", fn, "with a msg length of:", len(msgTexts), "and a result length of:", len(res))
}

func doMsgDecode(msgBytes []byte, fileName string) string {
	if len(msgBytes) <= 3 {
		return ""
	}

	msg := Message{}
	err := json.Unmarshal(msgBytes, &msg)
	if err != nil {
		log.Println("Error found when unmarshaling JSON in file:", fileName, "Error:", err, "Msg:", string(msgBytes))
		return ""
	}

	if msg.Event != "message" || msg.Text == "" {
		return ""
	}

	return strings.Replace(msg.Text, "\n", "\\n", -1)
}
