package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/gen2brain/beeep"
	"github.com/gocolly/colly"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var articleIds []int
var articleTitles []string
var notFirstTurn = false

func main() {

	alert("Starting scraper", "Launching..")

	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for t := range ticker.C {
			fmt.Println("Tick at", t)

			readWebsite("https://www.lemonde.fr/", "span[class=article__title-label]", notFirstTurn)
		}
	}()

	time.Sleep(720 * time.Hour)
	ticker.Stop()

	alert("Shutdown", "This program is running since 720h, 1 month, and will stop. Take a rest.")

	fmt.Println("hello world")
}

func alert(title string, bodyMessage string) {

	alert := beeep.Alert(title, bodyMessage, "assets/icongo.png") // The Gopher Mascot

	if alert != nil {
		panic(alert)
	}
}

func readWebsite(targetURL string, targetHTML string, activeAlert bool) {
	c := colly.NewCollector()

	c.OnHTML(targetHTML, func(e *colly.HTMLElement) {

		var title = e.Text
		var id = createID(title)

		if !ContainsInt(articleIds, id) {
			articleTitles = append(articleTitles, title)
			articleIds = append(articleIds, id)

			if activeAlert {
				alert("New article", tryCleanFormat(title))
			}
		}

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(targetURL)

	notFirstTurn = true
}

func createID(titleToConvert string) int {

	if len(titleToConvert) == 0 {
		return 0
	}

	var createdID int

	for pos, char := range titleToConvert {
		createdID += int(char) + pos
	}

	return createdID
}

func ContainsInt(ids []int, value int) bool {

	set := make(map[int]bool)
	for _, v := range ids {
		set[v] = true
	}

	return set[value]
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

/// ToDo find the right way without breaking the sentence
func tryCleanFormat(title string) string {

	title = strings.ReplaceAll(title, "é", "e")
	title = strings.ReplaceAll(title, "à", "a")
	title = strings.ReplaceAll(title, "'", " ")
	title = strings.ReplaceAll(title, "’", " ")
	title = strings.ReplaceAll(title, "«", " ")
	title = strings.ReplaceAll(title, "»", " ")

	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	titleNormalized, _, _ := transform.String(t, title)

	/*reg, err := regexp.Compile("[^a-zA-Z0-9 ]+ ")
	if err != nil {
		fmt.Println(err)
	}
	titleNormalized = reg.ReplaceAllString(title, "")*/
	fmt.Println(title + " ||| " + titleNormalized)

	return titleNormalized
}
