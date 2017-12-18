package craw

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"app/db"
	"app/utils"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/araddon/dateparse"
	"github.com/yhat/scrape"
)

func mockGetUrl() string {
	return "https://transparencia.campogrande.ms.gov.br/servidores/?action=get_simplificado_JSON&ajax=true&ano=2017&controller=servidores&length=8000&mes=10&orgao=8&start=0"
}

type Employee struct {
	Draw            string `json:"draw"`
	RecordsTotal    int    `json:"recordsTotal"`
	RecordsFiltered int    `json:"recordsFiltered"`
	Data            [][]string
}

type Payment struct {
	Salary   float64
	Discount float64
}

func StartCampoGrandeCrawler() {
	url := getAllUrl()
	for i := 0; i < len(url); i++ {
		resp, err := http.Get(url[i])
		if err != nil {
			fmt.Println("ERROR TO GET PAGE")
			fmt.Println(err)
		}
		empl := new(Employee)
		json.NewDecoder(resp.Body).Decode(&empl)
		findAllEmployees(empl.Data)
	}
}

func getAllUrl() []string {

	const AGENCY = 14

	urlArr := []string{}
	for i := 1; i <= AGENCY; i++ {
		url := getUrlCG()
		q := url.Query()
		q.Set("orgao", strconv.Itoa(i))
		url.RawQuery = q.Encode()
		urlArr = append(urlArr, url.String())
	}
	return urlArr
}

func getUrlCG() *url.URL {

	initUrl := new(url.URL)
	initUrl.Scheme = "https"
	initUrl.Host = "transparencia.campogrande.ms.gov.br"
	initUrl.Path = "servidores/"
	q := initUrl.Query()
	q.Set("controller", "servidores")
	initUrl.RawQuery = q.Encode()
	q.Set("action", "get_simplificado_JSON")
	q.Set("ano", "2017")
	q.Set("mes", "10")
	q.Set("ajax", "true")
	q.Set("start", "0")
	q.Set("length", "8000")

	initUrl.RawQuery = q.Encode()

	return initUrl
}

func findAllEmployees(emp [][]string) {

	for i := 0; i < len(emp); i++ {
		w := new(utils.Worker)
		w.Name = emp[i][0]
		w.Job = emp[i][3]

		date, err := convertToTime(emp[i][1])
		if err != nil {
			fmt.Println(err)
		}
		w.Date = date

		salaryUrl := emp[i][6]
		s, err := getSalary(salaryUrl)
		if err != nil {
			fmt.Println(err)
		}
		w.Salary = s.Salary

		err = persistWorker(w)
		if err != nil {
			fmt.Println("ERROR PERSIST WORKER: ", w)
			fmt.Println(err)
		}
	}
}

func persistWorker(w *utils.Worker) error {
	err := db.InsertPerson(w)
	if err != nil {
		return err
	}
	return nil
}

func convertToTime(value string) (t time.Time, err error) {

	date := strings.Replace(value, "/", "/25/", -1)
	dateTime, err := dateparse.ParseLocal(date)
	if err != nil {
		return dateTime, err
	}
	return dateTime, nil
}

func getSalary(param string) (p *Payment, err error) {

	salaryUrl := getSalaryUrl(param)

	resp, err := http.Get(salaryUrl)
	if err != nil {
		return p, err
	}
	salaryPage, err := html.Parse(resp.Body)

	if err != nil {
		return p, err
	}
	salaryDirt := getSalaryFromHtml(salaryPage)

	pay, err := arrangeSalary(salaryDirt)

	if err != nil {
		return p, err
	}
	return pay, nil
}

func arrangeSalary(value []string) (pay *Payment, err error) {

	p := new(Payment)
	p.Salary, err = arrangeSalaryString(value[0])
	if err != nil {
		return p, err
	}
	p.Discount, err = arrangeSalaryString(value[1])
	if err != nil {
		return p, err
	}
	return p, nil
}

func arrangeSalaryString(value string) (val float64, err error) {

	sal := strings.Replace(value, "R$ ", "", -1)
	sal = strings.Replace(sal, ".", "", -1)
	sal = strings.Replace(sal, ",", ".", -1)

	f, err := strconv.ParseFloat(sal, 2)
	if err != nil {
		return -1, err
	}
	return f, nil
}

func getSalaryFromHtml(page *html.Node) []string {

	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Tfoot {
			return true
		}
		return false
	}

	values := scrape.FindAll(page, matcher)
	salary := []string{}
	for _, salaryNode := range values {
		salary = append(salary, getSalaryValue(salaryNode))
	}
	return salary
}

func getSalaryValue(val *html.Node) string {

	matcher := func(n *html.Node) bool {
		if scrape.Attr(n, "class") == "text-right" {
			return true
		}
		return false
	}
	salary, _ := scrape.Find(val, matcher)
	return scrape.Text(salary)
}

func getSalaryUrl(param string) string {
	return strings.Split(param, "\"")[1]
}
