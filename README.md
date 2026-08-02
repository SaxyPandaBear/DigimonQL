# DigimonQL

[![Build Passing](https://github.com/SaxyPandaBear/DigimonQL/actions/workflows/ci.yaml/badge.svg)](https://github.com/SaxyPandaBear/DigimonQL/actions/workflows/ci.yaml)
[![API Functional](https://github.com/SaxyPandaBear/DigimonQL/actions/workflows/post_deploy.yaml/badge.svg)](https://github.com/SaxyPandaBear/DigimonQL/actions/workflows/post_deploy.yaml)

Inspired by [PokeApi](https://pokeapi.co/), with the dream of being as comprehensive, despite Digimon information being pretty scattered.

Main source of truth is the [Digimon Reference Book](https://digimon.net/reference_en/), but the annoying thing about the data is that their identifiers are inconsistent, e.g.: `rosemonburstmode` for Rosemon's Burst Mode form compared to `armamon_burstmode` for Armamon, and `miragegaogamon:burstmode` for MirageGaogamon. There's also the messy business of handling the English localizations, e.g.: `Diablomon` becomes `Diaboromon`. 

My hope is to expose an API that is easy to operate on, vetted against good source data, so that the Digimon community can flourish. The intent of this project is *not* to build a repository for the Digimon TCG - that already exists. 

Note: It is an intentional design choice because of the one-to-many nature of digivolutions to not implement a nested model.
Had it been done that way, the complexity of the return value would create too much overhead because of the branching.

## Usage
TBD

## Running locally

### Docker Compose
Prerequisite is to have the JSON data stored in `./data/digimon.json`, which is the output from the scraper. 

```bash
docker compose up --build
```

This should bring up the seeded MongoDB instance and the API. The API comes prepackaged with a GraphiQL visualizer.

You can connect to the local MongoDB instance on port `27017`, and the API is exposed on port `8081`.

Verify that the API is up by making a GraphQL query against it:
```graphql
query Digimon {
    digimon(id: "agumon") {
        name
        level
    }
}
```

#### GraphiQL in-browser explorer
![GraphiQL](./docs/graphiql_demo.png)

#### API call via Postman
![Postman-API-Call](./docs/postman.png)

### Without Docker

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

Run the server, backed directly by the local JSON file stored at `./data/digimon.json`:
```bash
go run server.go
```

### Testing

#### Local tests
```bash
go test ./... -v -cover
```

Note: There are some flaky tests around translating a Go struct into a BSON document. 
Not sure what the issue is because it's inconsistent. Try to `go clean` and retry, but
if it's a blocking issue with running tests, those can specifically be skipped by running:
```bash
go test ./... -short
```
This is how the CI is configured.

#### Integ tests
There are integration/e2e tests in `./scraper/smoke_test.py` that can be run in the virtual environment.

First, source the virtual environment.
```bash
cd ./scraper && source bin/activate
```

Then run it (this expects an `API_BASE_URL` environment variable to be set to the base URL of the service):
```bash
python smoke_test.py
```

### Scraping the data

#### Digimon Info
* [Digimon Reference Book](https://digimon.net/reference_en/)

#### Evolution Info
The evolution information that is written in `./scraper/evolutions.py` is painstakingly handwritten by me. 
No AI could do this work. I have to map digivolutions across canons, dropping the ones that aren't present
in the Encyclopedia, and also translating some from their English localized names in order to properly reference
them.
For this, I am excluding (to my best ability) pendulum evolutions and most warp evolutions.
I am also ignoring X Antibody characters for the time being.

Wikimon as a source includes every possible evolution, so it is feasible to scrape
that site to generate the evolution mappings. That being said, some of them are
technically correct but I just personally disagree. For example, for the protagonist
partners in Digimon Beatbreak, they list the Ultimate forms as warp evolutions: https://wikimon.net/Scourge_Chiropmon

* [Digimon Story: Time Stranger](https://www.grindosaur.com/en/games/digimon-story-time-stranger/digimon)
    * Completed ✅
* [Digimon World: Next Order](https://www.grindosaur.com/en/games/digimon-world-next-order/digimon)
    * Completed ✅
* [Digimon Story: Cyber Sleuth](https://www.grindosaur.com/en/games/digimon-story-cyber-sleuth/digimon)
    * Completed ✅
* [Wikimon](https://wikimon.net/) to fill the gaps

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
mongoimport --jsonArray --authenticationDatabase=admin -d public -c digimon --drop mongodb://something ./data/digimon.json
```
This example includes the `--drop` flag in order to completely refresh the collection. Not sure if there's
a clean way to do full upserts of the database.
