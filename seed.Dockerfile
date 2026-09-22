# This is for local development against MongoDB, so I don't care about the exposed URL
FROM mongo

COPY data/digimon.json /digimon.json
CMD mongoimport --authenticationDatabase=admin -d public -c digimon --type json --file /digimon.json --drop --jsonArray mongodb://root:example@mongo:27017/
