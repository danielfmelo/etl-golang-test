package craw

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app/utils"
	"app/db"

	"github.com/extrame/xls"
	"github.com/rs/zerolog/log"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func StartBHCrawler() {
	log.Info().
		Msgf("STARTING BELO HORIZONTE CRAWLER")

	url := getUrlBH()
	resp, err := http.Get(url)
	if err != nil {
		log.Info().
			Err(err).
			Msgf("ERROR TO GET PAGE")
	}
	page, err := html.Parse(resp.Body)
	if err != nil {
		log.Info().
			Err(err).
			Msgf("ERROR TO PARSE HTML PAGE")
	}

	mainDiv := getMainDiv(page)
	spreads := getSpreadsheetAddress(mainDiv)

	err = extractXls(spreads[0])

	if err != nil {
		log.Info().
			Err(err).
			Msgf("ERROR TO EXTRACT XLS")
	}
	log.Info().
		Msgf("BELO HORIZONTE CRAWLER FINISHED")
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

	log.Info().
		Msgf("Status: Getting latest XLS reference")

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
	log.Info().
		Msgf("Status: Downloading XLS file")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	spreadBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Info().
			Err(err).
			Msgf("ERROR TO TRANSFORM READER TO BYTES")
	}
	log.Info().
		Msgf("Status: Loading XLS file as Reader")
	spreadRead := bytes.NewReader(spreadBytes)

	file, err := xls.OpenReader(spreadRead, "uft-8")
	if err != nil {
		log.Info().
			Err(err).
			Msgf("ERROR TO OPEN READER")
	}
	log.Info().
		Msgf("Status: Reading information, transforming and saving in Database")
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
			err = db.InsertPerson(w)
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

	dateTime, err := utils.ConvertToTime(date)
	if err != nil {
		log.Info().
			Err(err).
			Msgf("ERROR TO CONVERT TO TIME")
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
