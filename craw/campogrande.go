package craw

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"app/db"
	"app/utils"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/rs/zerolog/log"
	"github.com/yhat/scrape"
)

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

//Start crawler
//Get URLs for each agency (orgao)
//For each URL, find all employees, transform the information and save in database
func StartCampoGrandeCrawler() {
	log.Info().
		Msgf("STARTING CAMPO GRANDE CRAWLER")

	url := getAllUrl()
	for i := 0; i < len(url); i++ {
		log.Info().
			Msgf("Status: Reading information from AGENCY: %d, transforming and saving in Database", i+1)
		resp, err := http.Get(url[i])
		if err != nil {
			log.Info().
				Err(err).
				Msgf("ERROR TO GET PAGE")
		}
		empl := new(Employee)
		json.NewDecoder(resp.Body).Decode(&empl)
		findAllEmployees(empl.Data)
	}
	log.Info().
		Msgf("CAMPO GRANDE CRAWLER FINISHED")
}

//Return an array of URLs changing the "orgao" parameter
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

//Create and return the URL
//TODO: change to get for other years and months
//TODO: change to use the correct length of the array
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
	q.Set("length", "80000")

	initUrl.RawQuery = q.Encode()

	return initUrl
}

//Transform and save information from each employee
func findAllEmployees(emp [][]string) {

	for i := 0; i < len(emp); i++ {
		w := new(utils.Worker)
		w.Name = emp[i][0]
		w.Job = emp[i][3]

		date, err := utils.ConvertToTime(emp[i][1])
		if err != nil {
			log.Info().
				Err(err).
				Msgf("ERROR TO CONVERT TO TIME")
		}
		w.Date = date

		salaryUrl := emp[i][6]
		s, err := getSalary(salaryUrl)
		if err != nil {
			log.Info().
				Err(err).
				Msgf("ERROR TO GET SALARY")
		}
		w.Salary = s.Salary

		err = db.InsertPerson(w)
		if err != nil {
			log.Info().
				Err(err).
				Msgf("ERROR TO PERSIST WORKER: ", w)
		}
	}
}

//Get total salary and the discounts (not stored)
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

//Transform and save the salary in a struct
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

//Transform the salary information in float
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

//Scrape salary from HTML
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

//Scrape salary from HTML
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

//Salary URL
func getSalaryUrl(param string) string {
	return strings.Split(param, "\"")[1]
}
