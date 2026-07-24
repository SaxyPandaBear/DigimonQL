"""
The main point of compexlity with digivolutions is that one digimon
can evolve into many different forms, and there are shared forms.
For example, both Piyomon and Coronamon (very different) are able
to digivolve into Birdramon in the Time Stranger game. And then
Birdramon can digivolve into different forms, e.g.: Garudamon and Parrotmon.
There's no programmatic way to derive these evolution chains, and the
official reference doesn't track it. The chains need to be written by hand,
and unlike something like Pokemon, the evolution chains are nonunique.
"""

# track a digimon and the set of digimon that they can evolve into
next_evolutions = {
    "omegamon": ["omegamon_merciful", "imperialdramonpaladinmode"],
    "imperialdramonfightermode": ["imperialdramonpaladinmode"],
    "lucemon": ["rucemonfalldownmode"],
    "rucemonfalldownmode": ["lucemonsatanmode", "lucemonlarva"],
    "alphamon": ["alphamon:ouryuken"],
    "ouryumon": ["alphamon:ouryuken"]
}

# track a digimon and the set of digimon that they can come from
previous_evolutions = {
    "rucemonfalldownmode": ["lucemon"],
    "lucemonsatanmode": ["rucemonfalldownmode"],
    "lucemonlarva": ["rucemonfalldownmode"],
    "alphamon:ouryuken": ["alphamon", "ouryumon"]
}
