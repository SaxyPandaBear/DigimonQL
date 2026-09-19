# Architecture

Infrastructure is hosted on [Railway](https://railway.com/), project `DigimonQL`, single `production` environment (region `us-east4-eqdc4a`).

```mermaid
flowchart LR
    subgraph External["External"]
        Client["API Consumer<br/>(GraphQL client)"]
        DigimonNet["digimon.net<br/>(reference site)"]
    end

    subgraph Pipeline["Data Pipeline — DigivolutionScraper (Airflow)<br/>repo: SaxyPandaBear/DigivolutionScraper"]
        Airflow["Airflow scheduler/webserver<br/>DAGs: scrape → transform → load"]
    end

    subgraph API["Main GraphQL API — DigimonQL (this repo)<br/>Go: gin + gqlgen"]
        Gin["gin HTTP server :8080<br/>GET / → GraphQL Playground<br/>POST /query → GraphQL handler"]
        Resolvers["graph/ resolvers<br/>(schema.resolvers.go)"]
        RepoLayer["db.DigimonRepository<br/>MongoDBRepository (or LocalDigimonRepository<br/>fallback via bundled data/digimon.json)"]
        RateLimit["limiter/<br/>• AuthenticatedRateLimitHandler (gin middleware)<br/>• CountQueryLimitHandler (GraphQL op middleware,<br/>  sliding-window guard on the Count query)"]
    end

    subgraph Data["Data Stores"]
        Mongo[("MongoDB 8.3.7<br/>digimon documents<br/>vol: mongodb-volume 5GB")]
        Redis[("Redis 8.10<br/>rate-limit counters<br/>vol: redis-volume 5GB, AOF")]
        Postgres[("Postgres 18 (ssl)<br/>Airflow metadata DB only<br/>vol: postgres-volume 5GB")]
    end

    Client -->|"HTTPS<br/>digimonql-production.up.railway.app"| Gin
    Gin --> Resolvers --> RepoLayer
    Gin --> RateLimit
    RateLimit -->|"REDIS_URL (private net)"| Redis
    RepoLayer -->|"MONGO_URL (private net)"| Mongo

    Airflow -->|scrapes| DigimonNet
    Airflow -->|"MONGO_URI (private net)<br/>load scraped Digimon docs"| Mongo
    Airflow -->|"AIRFLOW__DATABASE__SQL_ALCHEMY_CONN<br/>(private net)"| Postgres

    Mongo -.->|"public TCP proxy<br/>tokaido.proxy.rlwy.net:51634<br/>(ad-hoc/debug access)"| DevAccess["Developer / tooling"]
```

## Components

| Service | Role | Key detail |
|---|---|---|
| **DigimonQL** | Main GraphQL API | Go, `gin` + `gqlgen`. Public domain `digimonql-production.up.railway.app:8080`. Exposes `POST /query` and `GET /` (Playground). Connects to Mongo (`MONGO_URL`) and Redis (`REDIS_URL`). If `MONGO_URL` is unset it falls back to the bundled static `data/digimon.json` via `LocalDigimonRepository` — a local-dev/offline mode, not used in production since Mongo is configured. |
| **DigivolutionScraper** | Data pipeline | Separate repo, runs **Apache Airflow** (`AIRFLOW__CORE__*` vars) to orchestrate scraping [digimon.net](https://digimon.net/reference_en/) and loading results into MongoDB (`MONGO_URI`). This is the production analog of the `scraper/scrape.py` script checked into this repo (which only writes a local JSON file for dev use — see [Scraping the data](./README.md#scraping-the-data)). |
| **MongoDB** | Primary datastore | Holds the Digimon documents (name, level, type, attribute, moves, digivolution chains). Written by the scraper pipeline, read by the API. Also has a public TCP proxy for manual/dev inspection, outside of app traffic. |
| **Redis** | Rate-limiting store | Backs two layers in the API: a per-client bearer-token-aware limiter (`AuthenticatedRateLimitHandler`, gin middleware) and a sliding-window global limiter specifically on the expensive GraphQL `Count` operation (`CountQueryLimitHandler`). Private network only. |
| **Postgres** | Airflow metadata DB | Consumed **only** by DigivolutionScraper (`AIRFLOW__DATABASE__SQL_ALCHEMY_CONN`) as Airflow's internal scheduler/task-state store. The GraphQL API has no Postgres connection at all. Private network only, no public domain. |

All inter-service traffic uses Railway's private network (`*.railway.internal`, resolved via each service's private network endpoint: `digimonql`, `digivolutionscraper`, `mongodb`, `redis`, `postgres`), except the two public entry points: the DigimonQL API domain and Mongo's debug TCP proxy.
