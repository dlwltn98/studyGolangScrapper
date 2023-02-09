package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	title    string
	company  string
	category string
}

// Scrape Indeed by a term
func Scrape(term string) {
	var baseURL string = "https://search.incruit.com/list/search.asp?col=job&kw=" + term
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPage := getPages(baseURL)

	for i := 0; i < totalPage; i++ {
		go getPage(i, baseURL, c)
	}

	for i := 0; i < totalPage; i++ {
		extractJobs := <-c
		jobs = append(jobs, extractJobs...)
	}

	writeJobs(jobs)
	fmt.Println("Done extracted", len(jobs))
}

// 각 페이지별 데이터 추출 반환
func getPage(page int, url string, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := url + "&startno=" + strconv.Itoa(page*30)
	fmt.Println("Requesting", pageURL)

	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	// job card 조회
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".c_row")
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

// job card 데이터 추출
func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("jobno")
	title := CleanString(card.Find(".cell_mid .cl_top>a").Text())
	company := CleanString(card.Find(".cell_first .cl_top>a").Text())
	category := CleanString(card.Find(".cell_mid .cl_btm>span").Text())
	c <- extractedJob{
		id:       id,
		title:    title,
		company:  company,
		category: category}
}

// 페이지가 몇까지 있는지 알려줌
func getPages(url string) int {
	pages := 0
	res, err := http.Get(url)

	checkErr(err)
	checkCode(res)

	// res.Body는 byte, 입력/출력(io) → 닫아줘야함
	// 메모리가 새어나가는걸 막을 수 있음
	defer res.Body.Close()

	// Load the HTML doc
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".sqr_paging").Each(func(i int, s *goquery.Selection) {
		// i : 찾아내는 모든 아이템에 대한 것, s : selection
		pages = s.Find("a").Length()
	})

	// 오른쪽으로 이동하는 버튼이( > , >> ) 2개라 -1 함
	return pages - 1
}

// csv 파일에 저장
func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file) // 파일 새로 생성
	defer w.Flush()          // 파일에 데이터 입력

	// header 입력
	headers := []string{"ID", "Title", "Company", "Category"}
	wErr := w.Write(headers)
	checkErr(wErr)

	// 내용 입력
	for _, job := range jobs {
		jobSlice := []string{"https://job.incruit.com/jobdb_info/jobpost.asp?job=" + job.id, job.title, job.company, job.category}
		wErr := w.Write(jobSlice)
		checkErr(wErr)
	}
}

// 오류 처리
func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// 응답 코드 확인
func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}

// 공백 제거
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
