package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms/openai"
	"go.uber.org/zap"
)

func getLLM() (*openai.LLM, error) {
	// Replace this with your LLM url
	llmURL := "http://172.29.80.1:1234/v1/"
	llm, err := openai.New(openai.WithBaseURL(llmURL), openai.WithToken("lm-studio"))
	if err != nil {
		log.Error("openai.New failed", zap.Error(err))
	}
	return llm, err
}

func buildPrompt(verse Verse) string {
	prompt := "Recite the following bible verse. King James Version. Include only the exact text, do not add quotes or add anything extra."
	prompt += "\n"
	prompt += fmt.Sprintf("Book: %v, Chapter: %v, Verse: %v", verse.Book, verse.Chapter, verse.Verse)
	return prompt
}

func main() {
	llm, err := getLLM()
	if err != nil {
		log.Fatal("getLLM failed", zap.Error(err))
		return
	}

	// Read data from file
	data, err := os.ReadFile("./kjvdat.txt")
	if err != nil {
		log.Fatal("Error reading file", zap.Error(err))
		return
	}

	verses := parseVerses(string(data))
	results := compareVerses(verses, llm)
	log.Info("Test Finished", zap.Any("Results", results))
	log.Sugar().Infof("Accuracy: %v%%", getResultPercent(results)*100)
	csvOutput(results)
}

func csvOutput(results []LLMTestResult) {
	fmt.Println()
	curChapter := -1
	for _, result := range results {
		if curChapter != result.Verse.Chapter {
			if curChapter != -1 {
				fmt.Print("\n")
			}
			fmt.Print(result.Verse.Book + "," + strconv.Itoa(result.Verse.Chapter) + ",")
			curChapter = result.Verse.Chapter
		}
		value := 0
		if result.DoesMatch() {
			value = 1
		}
		fmt.Print(strconv.Itoa(value) + ",")
	}
	fmt.Println()
}

func getResultPercent(results []LLMTestResult) float64 {
	total := len(results)
	correct := 0
	for _, result := range results {
		if result.DoesMatch() {
			correct++
		}
	}
	return float64(correct) / float64(total)
}

func compareVerses(verses []Verse, llm *openai.LLM) []LLMTestResult {
	results := []LLMTestResult{}
	for _, verse := range verses {
		log.Sugar().Debugf("%v %v:%v", verse.Book, verse.Chapter, verse.Verse)
		prompt := buildPrompt(verse)
		ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
		result, err := llm.Call(ctx, prompt)
		if err != nil {
			log.Error("LLM call failed", zap.Error(err))
			result = ""
		}
		log.Debug(result)

		testResult := LLMTestResult{
			Verse:  verse,
			Actual: result,
		}
		results = append(results, testResult)

		match := result == verse.Line
		log.Sugar().Infof("%v", match)
		log.Info("RESULT", zap.Any("Book", verse.Book), zap.Any("Chapter", verse.Chapter), zap.Any("Verse", verse.Verse), zap.Any("Expected", verse.Line), zap.Any("Actual", result), zap.Any("Success", match))
	}
	return results
}

func parseVerses(data string) []Verse {
	log.Info("Parsing verses")

	lines := strings.Split(data, "\r\n")
	verses := make([]Verse, 0, len(lines))

	for _, line := range lines {
		if strings.Split(line, "|")[0] != "Pe2" {
			continue
		}

		log.Sugar().Debugf("Working on line %v", line)

		parts := strings.Split(line, "|")

		log.Sugar().Debugf("Split into [%v]", strings.Join(parts, ", "))

		bookPart := parts[0]

		log.Sugar().Debugf("Book: %v", bookPart)
		fullBookName, exists := BookNames[bookPart]
		if !exists {
			log.Sugar().Debugf("Book name %v does not exist, skipping", bookPart)
			continue
		}
		log.Sugar().Debugf("Parsed into: %v", fullBookName)

		chapterPart := parts[1]
		log.Sugar().Debugf("Chapter: %v", chapterPart)
		chapterNum := parseNumber(chapterPart)
		log.Sugar().Debugf("Prased into: %v", chapterNum)

		versePart := parts[2]
		log.Sugar().Debugf("Verse: %v", versePart)
		verseNum := parseNumber(versePart)
		log.Sugar().Debugf("Prased into: %v", verseNum)

		textPart := parts[3]
		log.Sugar().Debugf("Text: %v", textPart)
		textPart = strings.TrimSpace(textPart)
		textPart = strings.TrimRight(textPart, "~")
		log.Sugar().Debugf("Parsed into: %v", textPart)

		verse := Verse{
			Book:    fullBookName,
			Chapter: chapterNum,
			Verse:   verseNum,
			Line:    textPart,
		}
		verses = append(verses, verse)
	}
	return verses
}

func parseNumber(s string) int {
	var num int
	fmt.Sscanf(s, "%d", &num)
	return num
}

var log *zap.Logger

func init() {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(zap.DebugLevel)
	log, _ = config.Build()
}

// Represents a single verse
type Verse struct {
	Book    string
	Chapter int
	Verse   int
	Line    string
}

// Represents a single test result
type LLMTestResult struct {
	Verse  Verse  // Original verse tested against
	Actual string // Actual result from AI
}

func (T LLMTestResult) DoesMatch() bool {
	return T.Actual == T.Verse.Line
}

// Comment these out if you don't want them to be included
var BookNames = map[string]string{
	"Gen": "Genesis",
	"Exo": "Exodus",
	"Lev": "Leviticus",
	"Num": "Numbers",
	"Deu": "Deuteronomy",
	"Jos": "Joshua",
	"Jdg": "Judges",
	"Rut": "Ruth",
	"Sa1": "1 Samuel",
	"Sa2": "2 Samuel",
	"Kg1": "1 Kings",
	"Kg2": "2 Kings",
	"Ch1": "1 Chronicles",
	"Ch2": "2 Chronicles",
	"Ezr": "Ezra",
	"Neh": "Nehemiah",
	"Est": "Esther",
	"Job": "Job",
	"Psa": "Psalms",
	"Pro": "Proverbs",
	"Ecc": "Ecclesiastes",
	"Sol": "Song of Solomon",
	"Isa": "Isaiah",
	"Jer": "Jeremiah",
	"Lam": "Lamentations",
	"Eze": "Ezekiel",
	"Dan": "Daniel",
	"Hos": "Hosea",
	"Joe": "Joel",
	"Amo": "Amos",
	"Oba": "Obadiah",
	"Jon": "Jonah",
	"Mic": "Micah",
	"Nah": "Nahum",
	"Hab": "Habakkuk",
	"Zep": "Zephaniah",
	"Hag": "Haggai",
	"Zac": "Zechariah",
	"Mal": "Malachi",
	// "Es1": "1 Esdras",
	// "Es2": "2 Esdras",
	// "Tob": "Tobias",
	// "Jdt": "Judith",
	// "Aes": "Additions to Esther",
	// "Wis": "Wisdom",
	// "Bar": "Baruch",
	// "Epj": "Epistle of Jeremiah",
	// "Sus": "Susanna",
	// "Bel": "Bel and the Dragon",
	// "Man": "Prayer of Manasseh",
	// "Ma1": "1 Macabees",
	// "Ma2": "2 Macabees",
	// "Ma3": "3 Macabees",
	// "Ma4": "4 Macabees",
	// "Sir": "Sirach",
	// "Aza": "Prayer of Azariah",
	// "Lao": "Laodiceans",
	// "Jsb": "Joshua B",
	// "Jsa": "Joshua A",
	// "Jdb": "Judges B",
	// "Jda": "Judges A",
	// "Toa": "Tobit BA",
	// "Tos": "Tobit S",
	// "Pss": "Psalms of Solomon",
	// "Bet": "Bel and the Dragon Th",
	// "Dat": "Daniel Th",
	// "Sut": "Susanna Th",
	// "Ode": "Odes",
	"Mat": "Matthew",
	"Mar": "Mark",
	"Luk": "Luke",
	"Joh": "John",
	"Act": "Acts",
	"Rom": "Romans",
	"Co1": "1 Corinthians",
	"Co2": "2 Corinthians",
	"Gal": "Galatians",
	"Eph": "Ephesians",
	"Phi": "Philippians",
	"Col": "Colossians",
	"Th1": "1 Thessalonians",
	"Th2": "2 Thessalonians",
	"Ti1": "1 Timothy",
	"Ti2": "2 Timothy",
	"Tit": "Titus",
	"Plm": "Philemon",
	"Heb": "Hebrews",
	"Jam": "James",
	"Pe1": "1 Peter",
	"Pe2": "2 Peter",
	"Jo1": "1 John",
	"Jo2": "2 John",
	"Jo3": "3 John",
	"Jde": "Jude",
	"Rev": "Revelation",
}
