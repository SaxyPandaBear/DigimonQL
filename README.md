# DigimonQL

Inspired by [PokeApi](https://pokeapi.co/), with the dream of being as comprehensive, despite Digimon information being pretty scattered.

Main source of truth is the [Digimon Reference Book](https://digimon.net/reference_en/), but the annoying thing about the data is that their identifiers are inconsistent, e.g.: `rosemonburstmode` for Rosemon's Burst Mode form compared to `armamon_burstmode` for Armamon, and `miragegaogamon:burstmode` for MirageGaogamon. There's also the messy business of handling the English localizations, e.g.: `Diablomon` becomes `Diaboromon`. 

My hope is to expose an API that is easy to operate on, vetted against good source data, so that the Digimon community can flourish. The intent of this project is *not* to build a repository for the Digimon TCG - that already exists. 

## Usage
TBD

## Running locally

This uses [`gqlgen`](https://gqlgen.com/getting-started/) to generate the GraphQL models and plumbing,
and is served via [Gin](https://github.com/gin-gonic/gin) over HTTP.

Add the generator tool as a dependency:
```bash
go get -tool github.com/99designs/gqlgen
```

Generate the GraphQL models:
```bash
go tool gqlgen generate
```

Run the server:
```bash
go run server.go
```

### Scraping the data

IMPORTANT NOTE: Just running the scraper does not provide the full set of data right now, 
as things need to be added manually such as the mappings for evolutions and mode changes.

These steps assumes you already have a Python virtual environment configured in `./scraper`
```bash
cd scraper && source bin/activate
```

Install dependencies and execute the scraper:
```bash
pip install -r requirements.txt && python scrape.py
```

The script should output a JSON file to `../data/digimon.json`, which will then be used to serve data
in the GraphQL API.

#### Importing scraped data into MongoDB
The intent is to back the API with MongoDB documents. After installing `mongoimport`, you can directly
load the output JSON file into a collection. Note that out of the box, the JSON output from the scraper
is incomplete.

Load the data to your desired MongoDB instance:
```bash
mongoimport --jsonArray --authenticationDatabase=admin --drop mongodb://something ./data/digimon.json
```
This example includes the `--drop` flag in order to completely refresh the collection. Not sure if there's
a clean way to do full upserts of the database.
