package craw

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app/utils"

	"github.com/extrame/xls"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func StartBHCrawler() {
	url := getUrlBH()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("ERROR TO GET PAGE")
		fmt.Println(err)
	}
	page, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("ERROR TO PARSE HTML PAGE")
		fmt.Println(err)
	}

	mainDiv := getMainDiv(page)
	spreads := getSpreadsheetAddress(mainDiv)

	err = extractXls(spreads[0])

	if err != nil {
		fmt.Println("ERROR TO EXTRACT XLS")
		fmt.Println(err)
	}
}

func getUrlBH() string {
	url := "http://portalpbh.pbh.gov.br/pbh/ecp/comunidade.do?evento=portlet&pIdPlc=ecpTaxonomiaMenuPortal&app=acessoinformacao&lang=pt_BR&pg=10125&tax=41984"
	return url
}

func getParsedUrl() string {
	url := "http://portalpbh.pbh.gov.br"
	return url
}

func getMainDiv(page *html.Node) *html.Node {
	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Div && scrape.Attr(n, "id") == "PORTLET_CONTEUDO_0" {
			return true
		}
		return false
	}
	mainDiv, _ := scrape.Find(page, matcher)
	return mainDiv
}

func getSpreadsheetAddress(div *html.Node) []string {

	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.A {
			return true
		}
		return false
	}
	spread := scrape.FindAll(div, matcher)

	var spreads []string
	for _, s := range spread {
		spreads = append(spreads, getParsedUrl()+scrape.Attr(s, "href"))
	}
	return spreads
}

func extractXls(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	spreadBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	spreadRead := bytes.NewReader(spreadBytes)

	fmt.Printf("%T", spreadRead)

	file, err := xls.OpenReader(spreadRead, "uft-8")
	if err != nil {
		fmt.Println(err)
	}
	if sheet1 := file.GetSheet(0); sheet1 != nil {

		for i := 7; i <= (int(sheet1.MaxRow)); i++ {
			row := sheet1.Row(i)
			colName := row.Col(8)
			colSal := row.Col(11)
			colJob := row.Col(10)

			w := new(utils.Worker)
			w.Name = colName
			w.Job = colJob
			w.Salary, err = convertToFloat(colSal)
			if err != nil {
				return err
			}
			w.Date = getDateTime(url)
			err = persistWorker(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func convertToFloat(val string) (s float64, err error) {

	if val == "" {
		return 0.0, nil
	}
	f, err := strconv.ParseFloat(val, 2)
	if err != nil {
		return -1, err
	}
	return f, nil
}

func getDateTime(url string) time.Time {
	urlSplited := strings.Split(url, "_")
	month := urlSplited[len(urlSplited)-2]
	year := urlSplited[len(urlSplited)-1]
	year = strings.Replace(year, ".xls", "", -1)

	monthNumber := monthNameToDate(month)

	date := monthNumber + "/" + year

	dateTime, err := convertToTime(date)
	if err != nil {
		fmt.Println("ERROR TO CONVERT TO TIME ", err)
	}
	return dateTime
}

func monthNameToDate(name string) string {

	switch name {
	case "janeiro":
		return "1"
	case "fevereiro":
		return "2"
	case "marco":
		return "3"
	case "abril":
		return "4"
	case "maio":
		return "5"
	case "junho":
		return "6"
	case "julho":
		return "7"
	case "agosto":
		return "8"
	case "setembro":
		return "9"
	case "outubro":
		return "10"
	case "novembro":
		return "11"
	case "dezembro":
		return "12"
	default:
		return "0"
	}
}