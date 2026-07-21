# pyright: reportOptionalSubscript=false, reportOptionalMemberAccess=false
from bs4 import BeautifulSoup
from time import sleep
from reference import digimon_names, known_mode_variants

import json
import requests

output_path = "./digimon.json"
url_template = "https://digimon.net/reference_en/detail.php?directory_name="  # url param is the CASE SENSITIVE name of the digimon
img_domain = "https://digimon.net/"

# Tags to look up
parent_tag = "p-ref"  # encompassing class tag
en_name_tag = "c-titleSet__main"  # localized English name
info_tag = "p-ref__info"  # section that has details like level, type, attribute, and special move(s)
profile_tag = "p-ref__txt"  # description of the Digimon

# input is in the form <img src="../cimages/digimon/bearcatmon.jpg" alt="">
# so take that and replace the beginning with the domain.
def derive_img_src(src: str) -> str:
    return src.replace("../", img_domain)


# Example: "・Penetrate Blow・Murderize Rush・Beardown Spinning Kick"
# Calling split() will keep the empty string at the beginning, but for futureproofing,
# use a conditional list comprehension instead of just dropping the first element.
def parse_special_moves(s: str) -> list:
    return [move.strip() for move in s.strip().split("・") if len(move.strip()) > 0]


# The In-Training levels use the roman numerals I and II, in Unicode,
# but these aren't intuitively queryable compared to the numeric 1 and 2.
def clean_level(s: str) -> str:
    return s.replace("\u2160", "1").replace(u"\u2161", "2").replace("(Xros Wars)", "")


def clean_name(s: str) -> str:
    return s.replace("\uff1a", ":")


def clean_attribute(s: str) -> str:
    if len(s) == 0:
        return "None"
    return s

def main():
    # there should not be duplicates.
    # if there are duplicates, that has to be addressed before continuing
    # to scrape the data.
    name_set = set(digimon_names)
    diff = abs(len(digimon_names) - len(name_set))
    if diff != 0:
        print(f"Found {diff} duplicates in the data. Cannot proceed.")
        print([x for x in digimon_names if digimon_names.count(x) > 1])
        exit(1)

    data = []
    failures = []
    for name in digimon_names:
        digimon_url = f"{url_template}{name}"
        print(f"Checking {digimon_url}...")
        r = requests.get(digimon_url)
        soup = BeautifulSoup(r.text, features="html.parser")

        digimon = soup.find(class_=parent_tag)
        if digimon is None:
            print(f"Couldn't find data for {name} at {digimon_url}")
            print("Skipping to next digimon")
            failures.append(name)
            continue

        english_name = clean_name(digimon.find(class_=en_name_tag).text)
        img_url = derive_img_src(digimon.find("img")["src"])  # pyright:ignore

        info = digimon.find(class_=info_tag)
        if info is None:
            print(f"Couldn't find {info_tag} for {name}. Skipping.")
            failures.append(name)
            continue
        # There should be 4 elements: Level, Type, Attribute, Special Moves,
        # and the last element is a single string which may contain multiple values delimited by a dot character
        values = [t.text for t in info.find_all("dd")]
        digimon_level = clean_level(values[0])
        digimon_type = values[1]
        digimon_attr = clean_attribute(values[2])
        digimon_moves = parse_special_moves(values[3])

        result = dict()
        result["id"] = name  # identifier is the name used in the URL for the digimon
        result["name_en"] = english_name  # futureproofing this by suffixing with `en`
        result["level"] = digimon_level
        result["type"] = digimon_type
        result["attribute"] = clean_attribute(digimon_attr)
        result["moves"] = digimon_moves
        result["img_src"] = img_url
        result["background"] = digimon.find(class_=profile_tag).text.strip()
        result["previous_digivolutions"] = []  # not derivable from Reference Book, expected values are id values of other digimon
        result["next_digivolutions"] = []
        result["is_mode"] = name in known_mode_variants or " Mode" in english_name
        result["is_x_antibody"] = "(X Antibody)" in english_name

        data.append(result)

        sleep(0.25)  # wait for rate-limiting

    # after iterating over all of the digimon, check the failures (if any).
    # if there are failures, flag it to be addressed and exit early
    if len(failures) > 0:
        print(f"Had an issue finding {len(failures)} digimon. Triage these:")
        for name in failures:
            print(f"\t{name}")
        exit(1)

    # if there are no failures, write the data out as JSON to be used as the backing data for the database
    output = dict()
    output["digimon"] = data
    with open(output_path, "w") as f:
        json.dump(output, f)  # pyright:ignore
        print(f"Successfully wrote out {len(data)} digimon scraped to {output_path}")



if __name__ == "__main__":
    main()
