#!/bin/sh

echo "Waiting for Kafka Connect to start listening on kafka-connect ⏳"

while [ $(curl -s -o /dev/null -w %{http_code} http://localhost:8083/connectors) -eq 000 ] ; do 
  echo -e $(date) " Kafka Connect listener HTTP state: " $(curl -s -o /dev/null -w %{http_code} http://localhost:8083/connectors) " (waiting for 200)"
  sleep 5 
done

nc -vz localhost 8083

curl -i -X POST \
  -H "Accept:application/json" \
  -H "Content-Type:application/json" \
  http://localhost:8083/connectors/ \
  -d '{
  "name": "orders-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "plugin.name": "pgoutput",
    "tasks.max": "1",
    "database.hostname": "'$ORDER_DATABASE_HOST'",
    "database.port": '$ORDER_DATABASE_PORT',
    "database.user": "'$ORDER_DATABASE_USER'",
    "database.password": "'$ORDER_DATABASE_PASSWORD'",
    "database.dbname" : "'$ORDER_DATABASE_NAME'",
    "topic.prefix": "debezium",
    "heartbeat.interval.ms": "5000",
    "schema.include.list": "public",
    "table.include.list" : "public.*_events",
    "slot.name": "order_events_publication",
    "publication.name": "order_events_publication",
    "publication.autocreate.mode": "filtered",
    "transforms": "unwrap,PartitionRouting",
    "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
    "transforms.unwrap.add.fields": "op,table,lsn,source.ts_ms",
    "transforms.unwrap.delete.handling.mode": "rewrite",
    "transforms.unwrap.drop.tombstones": "true",
    "transforms.PartitionRouting.type": "io.debezium.transforms.partitions.PartitionRouting",
    "transforms.PartitionRouting.partition.payload.fields": "change.id",
    "transforms.PartitionRouting.partition.topic.num": 1,
    "message.key.columns": "public.order_create_events:id",
    "transforms.PartitionRouting.predicate": "allTopic",
    "predicates": "allTopic",
    "predicates.allTopic.type": "org.apache.kafka.connect.transforms.predicates.TopicNameMatches",
    "predicates.allTopic.pattern": ".*-events",
    "key.converter" : "org.apache.kafka.connect.storage.StringConverter",
    "key.converter.schemas.enable": false,
    "tombstones.on.delete": false,
    "null.handling.mode": "keep"
  }
}'

curl -i -X POST \
  -H "Accept:application/json" \
  -H "Content-Type:application/json" \
  http://localhost:8083/connectors/ \
  -d '{
  "name": "accounts-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "plugin.name": "pgoutput",
    "tasks.max": "1",
    "database.hostname": "'$ACCOUNT_DATABASE_HOST'",
    "database.port": '$ACCOUNT_DATABASE_PORT',
    "database.user": "'$ACCOUNT_DATABASE_USER'",
    "database.password": "'$ACCOUNT_DATABASE_PASSWORD'",
    "database.dbname" : "'$ACCOUNT_DATABASE_NAME'",
    "topic.prefix": "debezium",
    "heartbeat.interval.ms": "5000",
    "schema.include.list": "public",
    "table.include.list" : "public.*_events",
    "slot.name": "account_events_publication",
    "publication.name": "account_events_publication",
    "publication.autocreate.mode": "filtered",
    "transforms": "unwrap,PartitionRouting",
    "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
    "transforms.unwrap.add.fields": "op,table,lsn,source.ts_ms",
    "transforms.unwrap.delete.handling.mode": "rewrite",
    "transforms.unwrap.drop.tombstones": "true",
    "transforms.PartitionRouting.type": "io.debezium.transforms.partitions.PartitionRouting",
    "transforms.PartitionRouting.partition.payload.fields": "change.id",
    "transforms.PartitionRouting.partition.topic.num": 1,
    "message.key.columns": "public.account_events:id",
    "transforms.PartitionRouting.predicate": "allTopic",
    "predicates": "allTopic",
    "predicates.allTopic.type": "org.apache.kafka.connect.transforms.predicates.TopicNameMatches",
    "predicates.allTopic.pattern": ".*-events",
    "key.converter" : "org.apache.kafka.connect.storage.StringConverter",
    "key.converter.schemas.enable": false,
    "tombstones.on.delete": false,
    "null.handling.mode": "keep"
  }
}'

