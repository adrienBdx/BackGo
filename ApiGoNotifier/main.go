package main

import (
	"fmt"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/gocolly/colly"
)

var articleIds []int
var articleTitles []string

func alert(title string, bodyMessage string) {

	alert := beeep.Alert(title, bodyMessage, "assets/icongo.png")

	if alert != nil {
		panic(alert)
	}

	fmt.Println("hello world - " + title + ":" + bodyMessage)
}

func main() {

	alert("Starting scraper", "Launching..")

	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for t := range ticker.C {
			fmt.Println("Tick at", t)

			alert("Ticking", "First tick")
		}
	}()

	//time.Sleep(720 * time.Hour) ToDo Remove
	ticker.Stop()

	readWebsite()

	fmt.Println("test")
	/*
		c.Visit("https://www.boursorama.com/bourse/devises/taux-de-change-euro-dollarcanadien-EUR-CAD/")
		// span[c-instrument c-instrument--last]
		c.OnHTML("span[data-ist-last]", func(e *colly.HTMLElement) {
			fmt.Println(e.Attr("class"))
			fmt.Println(e.Text)
			fmt.Println(e.Name)
		})
	*/

	fmt.Println("hello world")

}

func readWebsite() {
	c := colly.NewCollector()
	c.OnHTML("span[class=article__title-label]", func(e *colly.HTMLElement) {
		fmt.Println(e.Attr("class"))
		fmt.Println(e.Text)
		fmt.Println(e.Name)
		fmt.Println(len(articleIds))

		articleTitles = append(articleTitles, e.Text)
		fmt.Println(len(articleTitles))
	})
	//LeMonde span[class=article__title-label]

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("https://www.lemonde.fr/")
}
