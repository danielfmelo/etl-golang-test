package main

import (
	"app/craw"
	"app/db"
)

func main() {

	db.InitDB()
	defer db.CloseDB()
	craw.StartBHCrawler()
	craw.StartCampoGrandeCrawler()
}
