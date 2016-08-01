package main

import (
	"flag"
	"io/ioutil"
	"strings"
	"fmt"
	"os"
	"sort"
)

var dataset = flag.String("dataset", "./dataset.txt", "the filepath to the dataset in plain text")
var out = flag.String("out", "./out.txt", "the output of the program")
var maxOutput = flag.Int("max", 10000, "the maximum words allowed to be outputted")

func main() {
	flag.Parse()

	fileBytes, err := ioutil.ReadFile(*dataset)
	if err != nil {
		panic(err)
	}


	fileStr := string(fileBytes)

	fileStr = strings.ToLower(fileStr)


	msgArr := strings.Split(fileStr, "\n")

	wordMap := map[string]int{
		".": 0,
		",": 0,
	}

	for _, msg := range msgArr {
		var startIndex int
		evalMsg := strings.Replace(msg, "\\n", "\n", -1)

		for i, char := range evalMsg {
			if char == ',' || char == '.' || char == ' ' || char == '\n' {
				if i != startIndex {
					tmp := evalMsg[startIndex:i]
					val, ok := wordMap[tmp]
					if ok {
						wordMap[tmp] = val + 1
					} else {
						wordMap[tmp] = 1
					}
				}
				startIndex = i + 1
				if char == ',' || char == '.' {
					strChar := string(char)
					val := wordMap[strChar]

					wordMap[strChar] = val + 1

				}
			}
		}
		if startIndex < len(evalMsg) {
			tmp := evalMsg[startIndex:]
			val, ok := wordMap[tmp]
			if ok {
				wordMap[tmp] = val + 1
			} else {
				wordMap[tmp] = 1
			}
		}
	}

	pl := rankByWordCount(wordMap)

	if len(pl) > *maxOutput {
		pl = pl[0 : 10000]
	}


	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	pl = append(PairList{Pair{"<UNK>", 0}}, pl...)

	for i, pair := range pl {
		_, err := f.WriteString(fmt.Sprintf("%d %s\n", i, pair.Key))
		if err != nil {
			panic(err)
		}
	}


}

func rankByWordCount(wordFrequencies map[string]int) PairList{
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{
			Key:k,
			Value: v,
		}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key string
	Value int
}

type PairList []Pair

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }