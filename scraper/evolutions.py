"""
The main point of compexlity with digivolutions is that one digimon
can evolve into many different forms, and there are shared forms.
For example, both Piyomon and Coronamon (very different) are able
to digivolve into Birdramon in the Time Stranger game. And then
Birdramon can digivolve into different forms, e.g.: Garudamon and Parrotmon.
There's no programmatic way to derive these evolution chains, and the
official reference doesn't track it. The chains need to be written by hand,
and unlike something like Pokemon, the evolution chains are nonunique.

Starting with Digimon Story: Time Stranger as a reference point for canon.
"""

# track a digimon and the set of digimon that they can evolve into
next_evolutions = {
    "bubbmon": ["pyocomon", "tanemon", "mochimon"],
    "dodomon": ["dorimon", "wanyamon"],
    "kuramon": ["pagumon", "tsumemon"],
    "punimon": ["tunomon", "nyaromon"],
    "botamon": ["koromon"],
    "poyomon": ["pukamon", "tokomon"],
    "choromon": ["capromon"],
    "tanemon": ["floramon", "funbeemon", "lalamon", "mushmon", "palmon"],
    "wanyamon": ["bakumon", "bearmon", "gaomon", "renamon"],
    "mochimon": ["tentomon", "gottsumon", "wormmon", "tyumon"],
    "pyocomon": ["piyomon", "hyokomon", "muchomon", "penmon", "falcomon", "hawkmon"],
    "tsumemon": ["gabumon_black", "shamamon", "dracumon", "keramon"],
    "pagumon": ["picodevimon", "gazimon", "impmon", "otamamon"],
    "dorimon": ["monodramon", "dorumon", "lopmon", "snowgoburimon"],
    "nyaromon": ["kudamon", "plotmon", "huckmon", "lunamon"],
    "algomon_lv1": ["algomon_lv2"],
    "algomon_lv2": ["algomon_lv3"],
    "algomon_lv3": ["algomon_lv4"],
    "algomon_lv4": ["algomon-ultimate"],
    "algomon-ultimate": ["algomon"],
    "omegamon": ["omegamon_merciful", "imperialdramonpaladinmode"],
    "imperialdramonfightermode": ["imperialdramonpaladinmode"],
    "lucemon": ["rucemonfalldownmode"],
    "rucemonfalldownmode": ["lucemonsatanmode", "lucemonlarva"],
    "alphamon": ["alphamon:ouryuken"],
    "ouryumon": ["alphamon:ouryuken"],
    "tunomon": ["zubamon", "ryudamon", "gabumon", "elecmon", "goburimon", "v-mon"],
    "koromon": ["dracomon", "kotemon", "agumon", "guilmon", "shoutmon", "betamon"],
    "pukamon": ["gomamon", "shakomon", "ganimon", "kamemon", "gizamon"],
    "tokomon": ["coronamon", "terriermon", "patamon", "armadimon", "lucemon"],
    "capromon": ["toyagumon", "kokuwamon", "solarmon", "hagurumon"],
    "floramon": ["sunflowmon", "kiwimon", "togemon", "woodmon", "vegimon"],
    "penmon": ["peckmon", "buraimon", "aquilamon", "kiwimon"],
    "wormmon": ["snimon", "dokugumon", "stingmon"],
    "funbeemon": ["kabuterimon", "dokugumon", "waspmon", "flymon", "stingmon", "goldnumemon"],
    "tentomon": ["kabuterimon", "snimon", "sunflowmon", "kuwagamon", "waspmon"],
    "bearmon": ["leomon", "gryzmon", "mojyamon", "gaogamon"],
    "coronamon": ["agnimon", "flaremon", "birdramon", "baohuckmon", "growmon", "meramon"],
    "lalamon": ["revolmon", "sunflowmon", "togemon", "turuiemon", "tuchidarumon"],
    "gaomon": ["leomon", "gaogamon", "turuiemon", "nanimon", "strikedramon"],
    "lopmon": ["leomon", "gryzmon", "minotaurmon", "turuiemon", "tuchidarumon", "wendimon"],
    "mushmon": ["flymon", "scumon", "nanimon", "woodmon"],
    "gazimon": ["dorugamon", "baohuckmon", "dobermon", "gaogamon", "sangloupmon", "blacktailmon"],
    "hyokomon": ["birdramon", "peckmon", "buraimon", "dinohumon"],
    "muchomon": ["birdramon", "airdramon", "fugamon", "peckmon"],
    "renamon": ["reppamon", "kyubimon", "sunflowmon", "lekismon"],
    "shamamon": ["witchmon", "minotaurmon", "fugamon", "musyamon"],
    "lunamon": ["lekismon", "hyougamon", "garurumon", "yukidarumon", "sorcerimon", "icemon"],
    "solarmon": ["starmon", "meramon", "goldnumemon", "gardromon_gold"],
    "gabumon_black": ["dobermon", "dorugamon", "blacktailmon", "fugamon", "minotaurmon", "garurumon_black"],
    "piyomon": ["yunimon", "wizarmon", "birdramon", "aquilamon"],
    "dracumon": ["starmon", "sangloupmon", "scumon", "wizarmon"],
    "picodevimon": ["greymon_blue", "devimon", "bakemon", "icedevimon", "orgemon"]
}
