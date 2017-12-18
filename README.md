# ETL Golang Test

A simple ETL project wrote in GOLANG. The goal of this project is test the skils in ETL.

## Getting Started

The project scrapes data from public sites and save in a PostgreSQL database. The information scraped is Name, Position and Salary from the employees of two brazilian prefectures.

The two prefectures fonts used were: 
* [PortalPBH](http://portalpbh.pbh.gov.br/pbh/ecp/comunidade.do?evento=portlet&pIdPlc=ecpTaxonomiaMenuPortal&app=acessoinformacao&lang=pt_BR&pg=10125&tax=41984) - Prefeitura Belo Horizonte
* [PortalPCG](https://transparencia.campogrande.ms.gov.br/servidores/) - Prefeitura Campo Grande

### Prerequisites

For run the project you must have the docker version >= 17.04 installed

```
$ docker -v
Docker version 17.04.0-ce, build 4845c56
```

### Setup

In the first time running the docker-compose, it will create the images and setting the database:

```
$ docker-compose run --service-ports app
```

After the first time it only will load the database, keeping the configuration.

Install all dependencies needed:

```
$ make deps
```

Run project:

```
$ make run
```

Format code:

```
$ make fmt
```

Clean project and packages:

```
$ make clean
```

Build project:

```
$ make build
```

## Dependencies

The golang packages used were:

* [Scrape](https://github.com/yhat/scrape) - An interface for Go web scraping.
* [Xls](https://github.com/extrame/xls) - Pure Golang xls library.
* [Dataparse](https://github.com/araddon/dateparse) - GoLang Parse any date string without knowing format in advance. 
* [pq](https://github.com/lib/pq) - Pure Go Postgres driver for database/sql. 
* [Zerolog](https://github.com/rs/zerolog/) - Logger dedicated to JSON output. 

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details