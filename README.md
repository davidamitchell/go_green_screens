
# Kafka notes

## Install
```sh
brew install kafka
```

## Start zookeeper
```sh
zookeeper-server-start /usr/local/etc/kafka/zookeeper.properties & kafka-server-start /usr/local/etc/kafka/server.properties
```

## Create a topic
```sh
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic some-topic
```

## List topics
```sh
kafka-topics --list --zookeeper localhost:2181
```

## Consume a topic
```sh
kafka-console-consumer --bootstrap-server localhost:9092 --topic some-topic --from-beginning
```

## Write a message to a topic
```sh
echo hi | kafka-console-producer --broker-list localhost:9092 --topic some-topic
```

## Needed env vars

```sh
export KAFKA_HOST=localhost
export KAFKA_PORT=9092
export DB_CONNECTION_STRING=dbname=greenscreen_dev user=765440 sslmode=disable
```
