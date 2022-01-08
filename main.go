package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
)

type scoreList struct {
	Objective string
	Lock      bool
	Score     int
	Name      string
}

var (
	filePath      = flag.String("file", "", "scoreboard.dat file Path")
	binary        = []byte{}
	scores        = []scoreList{}
	sortUpper     = flag.Bool("upper", true, "upper sort")      //昇順か否か
	searchKeyword = flag.String("key", "", "object search key") //正規表現(怪しいけど)
)

func main() {
	//flag 入手
	flag.Parse()
	handle, _ := os.Open(*filePath)
	defer handle.Close()
	zipReader, _ := gzip.NewReader(handle)
	defer zipReader.Close()
	binary, _ = ioutil.ReadAll(zipReader)

	i := 0

	for i < len(binary)-9 {
		//スタート入手
		if string(binary[i:i+9]) == "Objective" {
			//scoreのデータ
			ScoreData := scoreList{}
			// Objective nul ??で??からlength入手
			i = i + 10
			ObjectiveLengh, _ := strconv.Atoi(fmt.Sprint(binary[i]))
			i = i + 1
			ScoreData.Objective = string(binary[i : i+ObjectiveLengh])
			//Lockedのデータを入手
			i = i + ObjectiveLengh + 9
			ScoreData.Lock, _ = strconv.ParseBool(fmt.Sprint(binary[i]))
			//Scoreからデータを入手
			i = i + 9
			scoreA, _ := strconv.Atoi(fmt.Sprint(binary[i]))
			scoreB, _ := strconv.Atoi(fmt.Sprint(binary[i+1]))
			scoreC, _ := strconv.Atoi(fmt.Sprint(binary[i+2]))
			scoreD, _ := strconv.Atoi(fmt.Sprint(binary[i+3]))
			ScoreData.Score = scoreA*256*256*256 + scoreB*256*256 + scoreC*256 + scoreD
			//MSBが1のとき負の数字にする
			if scoreA > 127 {
				ScoreData.Score = ScoreData.Score - 256*256*256*256
			}
			//Nameを入手
			i = i + 4 + 8
			NameLengh, _ := strconv.Atoi(fmt.Sprint(binary[i]))
			i = i + 1
			ScoreData.Name = string(binary[i : i+NameLengh])
			n := i
			//Playerスコアか確認
			checkClear := true
			for n < i+NameLengh {
				byteData, _ := strconv.Atoi(fmt.Sprint(binary[n]))
				if byteData >= 0 && byteData <= 31 {
					checkClear = false
				}
				n++
			}
			if checkClear {
				scores = append(scores, ScoreData)
			}
		}
		i++
	}
	//ソート
	sort.SliceStable(
		scores,
		func(i, j int) bool {
			if *sortUpper {
				//昇順
				return scores[i].Score < scores[j].Score
			} else {
				//降順
				return scores[i].Score > scores[j].Score
			}
		},
	)

	for _, data := range scores {
		//検索
		if regexp.MustCompile(*searchKeyword).MatchString(data.Objective) {
			fmt.Println(data.Objective + "," + strconv.FormatBool(data.Lock) + "," + fmt.Sprint(data.Score) + "," + data.Name)
		}
	}
}
