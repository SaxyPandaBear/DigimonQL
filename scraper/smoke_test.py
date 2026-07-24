"""
Integration test script that should validate that the GraphQL API is working
as intended. This includes validating all of the query operations, and some
form of comprehensive testing on the digimons(input: Filter) query to ensure
that the filtering works as intended. There should also be an explicit validation
around querying for the `type` field of a Digimon document, because the naming
convention directly conflicts with a reserved word in GraphQL.
"""
import os
import requests

def validate_single_query(d: dict) -> bool:
    if d["digimonType"] != "Food":
        return False
    if d["name"] != "Burgamon":
        return False
    if d["level"] != "Rookie":
        return False
    if d["attribute"] != "Vaccine":
        return False
    return True

def validate_filter_query(digimons: list) -> bool:
    expected = {
        "shortmon": {
            "name": "Shortmon",
            "level": "Champion"
        },
        "burgamon_lv4": {
            "name": "Burgamon",
            "level": "Champion"
        },
        "torikaraballmon": {
            "name": "TorikaraBallmon",
            "level": "In-Training 2"
        },
        "potamon": {
            "name": "Potamon",
            "level": "Champion"
        },
        "burgamon": {
            "name": "Burgamon",
            "level": "Rookie"
        },
        "ebiburgamon": {
            "name": "EbiBurgamon",
            "level": "Rookie"
        }
    }

    if len(expected) != len(digimons):
        return False

    for d in digimons:
        if d["id"] not in expected:
            return False
        # everything in the filtered response should be of type "Food"
        if d["digimonType"] != "Food":
            return False
        comparison = expected[d["name"]]
        for k, v in comparison.items():
            if v != d[k]:
                return False
    return True

def validate_count_query(num: int) -> bool:
    return num >= 1314 # TODO: not sure if there's a good way to update this as new Digimon are added

def main():
    base_url = os.getenv("API_BASE_URL")
    if base_url is None:
        print("Failed to get base URL for the API to test against. Failing loudly...")
        exit(1)

    # create a multi-function query that tests all of the functionality in one request,
    # in order to minimize separate calls for network egress.
    """
    query IntegrationTest {
        digimon(id: "burgamon") {
            digimonType
            name
            level
            attribute
        }
        digimons(input: { digimonType: "Food" }) {
            name
            digimonType
            level
        }
        count
    }
    """
    # query_str has properly escaped quotes in the input
    query_str = "{\"query\": \"query IntegrationTest { digimon(id: \\\"burgamon\\\") { digimonType name level attribute } digimons(input: {digimonType: \\\"Food\\\"}) { name digimonType level } count}\"}"
    query = {
        "query": query_str
    }
    resp = requests.post(f"{base_url}/query", data=query, headers={"Content-Type": "application/json; charset=utf-8"})
    if resp.status_code != 200:
        print(f"Expected status code 200, but got {resp.status_code}")
        print(resp.json())
        exit(1)
    r = resp.json()
    if "errors" in r:
        print("There are errors returned from the API:")
        print(r["errors"])

    single_query = r["digimon"]
    filter_query = r["digimons"]
    count_query = r["count"]

    # if the code makes it this far, each of the queries returned something, so don't fail-fast here
    failed = False
    if not validate_single_query(single_query):
        print("Failed validation for single query.")
        print(single_query)
        failed = True
    if not validate_filter_query(filter_query):
        print("Failed validation for filter query.")
        print(filter_query)
        failed = True
    if not validate_count_query(count_query):
        print("Failed validation for count query.")
        print(count_query)
        failed = True
    if failed:
        exit(1)

    # successfully validated.
    print("Successfully validated API...")

if __name__ == "__main__":
    main()
