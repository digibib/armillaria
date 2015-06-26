sudo docker run -d -p 8890:8890 -p 1111:1111 -e DBA_PASSWORD=secret -e SPARQL_UPDATE=true --name=virtuoso tenforce/virtuoso /bin/bash /startup.sh
sudo docker run -d -p 9200:9200 -p 9300:9300 --name=elasticsearch elasticsearch:latest
cp data/config_example.json data/config.json
sudo docker build -t armillaria .
sleep 3
sudo docker run -d -p 8080:8080 -e SERVER_PORT=8080 --link elasticsearch:elasticsearch --link virtuoso:virtuoso --name=armillaria armillaria
sudo docker exec armillaria go run setupmappings.go indexing.go
