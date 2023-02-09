package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dlwltn98/learngo2/scrapper"
	"github.com/labstack/echo"
)

const FileName string = "jobs.csv"

func handleHome(c echo.Context) error {
	//return c.String(http.StatusOK, "Hello World")
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(FileName)
	fmt.Println(c.FormValue("term"))
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment(FileName, FileName) // 첨부파일 리턴
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}
