# Kafka Advanced Pub/Sub

Namespaces: `backend/pubsub/publisher/kafka` and `backend/pubsub/subscriber/kafka`

Use this instead of the basic `kafka://` driver when you need explicit broker connection settings, mTLS, or SASL.

## Capabilities

- mTLS with CA and client certificates
- SASL PLAIN and OAUTHBEARER authentication
- Azure Event Hub via `azure_event_hub: true` (SASL handshake v0)
- Producer `idempotent` mode (maps to `RequiredAcks=all`)
- Consumer group and isolation level configuration
- Message key from HTTP headers via `key_meta`
- Shared `cluster` config with the `async/kafka` agent driver

## Publisher

```json
{
  "host": ["ignore"],
  "url_pattern": "/ignore",
  "disable_host_sanitize": true,
  "extra_config": {
    "backend/pubsub/publisher/kafka": {
      "success_status_code": 201,
      "writer": {
        "topic": "events",
        "key_meta": "X-Event-Key",
        "cluster": {
          "brokers": ["localhost:9092"],
          "client_id": "pucora_publisher",
          "sasl": {
            "user": "user",
            "password": "password"
          }
        },
        "producer": {
          "idempotent": true
        }
      }
    }
  }
}
```

## Subscriber

```json
{
  "host": ["ignore"],
  "url_pattern": "/ignore",
  "disable_host_sanitize": true,
  "encoding": "json",
  "extra_config": {
    "backend/pubsub/subscriber/kafka": {
      "reader": {
        "topics": ["events"],
        "key_meta": "X-Event-Key",
        "cluster": {
          "brokers": ["localhost:9092"],
          "client_id": "pucora_subscriber"
        },
        "group": {
          "id": "my_group_id",
          "isolation_level": "read_commited"
        }
      }
    }
  }
}
```

## Azure Event Hub

Event Hubs require SASL handshake v0. Set `azure_event_hub: true` and use PLAIN with the connection string as the password:

```json
"cluster": {
  "brokers": ["my-namespace.servicebus.windows.net:9093"],
  "client_tls": {
    "allow_insecure_connections": false
  },
  "sasl": {
    "mechanism": "PLAIN",
    "user": "$ConnectionString",
    "password": "Endpoint=sb://...",
    "azure_event_hub": true
  }
}
```

If `user` is omitted with `azure_event_hub: true`, it defaults to `$ConnectionString`.

## OAUTHBEARER

Use `mechanism: "OAUTHBEARER"` and pass the bearer token in `password`:

```json
"sasl": {
  "mechanism": "OAUTHBEARER",
  "password": "<access-token>",
  "auth_identity": "optional-authzid"
}
```

## Offset commits

Kafka has no per-message ACK. Pucora commits partition offsets after each message is read, including when the decode step fails after a successful fetch.

## Example

```bash
make pubsub-kafka-advanced-compose-test
```

See [`examples/pubsub/kafka-advanced/`](../examples/pubsub/kafka-advanced/).

## Related

- [Basic Pub/Sub](pubsub.md)
- [Async Kafka agent](async-kafka.md)
