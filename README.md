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
TBD
