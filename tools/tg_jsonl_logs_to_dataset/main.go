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

// A Telegram message from the logs
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

// The result message from readLogFile()
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

	// A channel that is used for the output of readLogFile()
	outChan := make(chan ReadLogRes, len(logFiles))

	// Start reading log files
	for _, logFile := range logFiles {
		go readLogFile(outChan, logFile)
	}

	// Create an output file
	f, err := os.Create(*outFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Listen to the output channel from readLogFile() goroutines
	var numDone int
	for {
		res := <-outChan

		// Write to out file if the result was a success
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

		// If there are no more outputs, quit.
		numDone++
		if numDone >= len(logFiles) {
			break
		}
	}

	log.Println("Finished Converting Logs. Quitting...")
}

// This is run per log file. It gets a filename and a channel to send the result to. The result is a string of messages
// seperated by newline.
func readLogFile(outChan chan ReadLogRes, fn string) {
	var err error
	var res []byte

	defer func() {
		// All variables in here can be changed by the external prgm
		outChan <- ReadLogRes{
			Error:    err,
			Success:  err == nil && len(res) > 0,
			Output:   res,
			FileName: fn,
		}
	}()

	// Reads file to byte array
	fileBytes, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Println("Error! Could not read file:", fn, "With error:", err)
		return
	}

	// Split byte array by newline
	msgs := bytes.Split(fileBytes, []byte{'\n'})

	// Allocate a new array for the output of msgs
	msgTexts := make([]string, len(msgs))

	// Create a new sync
	wait := new(sync.WaitGroup)
	for i, msgBytes := range msgs {
		wait.Add(1)

		// Run msg decode and add the output to the new array
		go func(j int, msgB []byte, fileName string) {
			defer wait.Done()
			msgTxt := doMsgDecode(msgB, fileName)
			msgTexts[j] = msgTxt
		}(i, msgBytes, fn)
	}
	// Wait for the for loop data processing to finish
	wait.Wait()


	// Filters out blank messages
	filteredMsgTxts := msgTexts[:0]
	for _, x := range msgTexts {
		if x != "" {
			filteredMsgTxts = append(filteredMsgTxts, x)
		}
	}

	// Reverses messages so the newest ones are at the bottom
	for i := len(filteredMsgTxts)/2-1; i >= 0; i-- {
		opp := len(filteredMsgTxts)-1-i
		filteredMsgTxts[i], filteredMsgTxts[opp] = filteredMsgTxts[opp], filteredMsgTxts[i]
	}

	res = []byte(strings.Join(filteredMsgTxts, "\n"))
	log.Println("Finished", fn, "with a msg length of:", len(filteredMsgTxts), "and a result length of:", len(res))
}

// Reads given JSON string and outputs a sanitized text message or a blank string if a message is not available.
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
