# This is for local development against MongoDB, so I don't care about the exposed URL
FROM mongo

COPY data/digimon.json /digimon.json
CMD mongoimport --authenticationDatabase=admin -d public -c digimon --type json --file /digimon.json --drop --jsonArray mongodb://root:example@mongo:27017/ && mongo admin --authenticationDatabase=admin --host mongo -u root -p example --eval "db.createUser({user: 'tyk', pwd: 'tyk', roles: [{role: 'readWrite', db: 'tyk_analytics'}]}); db.createUser({user: 'tyk', pwd: 'tyk', roles: [{role: 'userAdminAnyDatabase', db: 'admin'}]});"
