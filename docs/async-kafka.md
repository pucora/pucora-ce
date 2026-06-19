# Async Kafka Agent

Namespace: `async/kafka` inside `async_agent[].extra_config`

Connects a background Pucora agent to Kafka topics for autonomous event consumption. Unlike endpoint Kafka subscribers, no HTTP client triggers processing.

## How it works

```
Kafka topic -> async/kafka driver -> internal proxy pipe -> backend webhook(s)
```

The topic name comes from `async_agent[].consumer.topic`. The driver uses a **latest offset** policy: only messages produced after the agent starts are consumed.

## Key capabilities

- mTLS and SASL PLAIN / OAUTHBEARER authentication (shared `cluster` config with Kafka Advanced Pub/Sub)
- Azure Event Hub via `sasl.azure_event_hub: true` (SASL handshake v0)
- Consumer group management (`group.id` or `group.group_id`)
- Message key exposed as HTTP header via `key_meta`
- Configurable isolation level
- Worker pool via `consumer.workers`
- Optional rate limiting via `consumer.max_rate`
- Health pings to the gateway `__health` endpoint

## Configuration

```json
{
  "version": 3,
  "async_agent": [
    {
      "name": "events-consumer",
      "consumer": {
        "topic": "events",
        "workers": 2,
        "timeout": "2s"
      },
      "backend": [
        {
          "host": ["http://analytics:8080"],
          "url_pattern": "/ingest",
          "method": "POST"
        }
      ],
      "connection": {
        "max_retries": 3,
        "backoff_strategy": "linear",
        "health_interval": "30s"
      },
      "extra_config": {
        "async/kafka": {
          "cluster": {
            "brokers": ["localhost:9092"],
            "client_id": "pucora_async_agent"
          },
          "group": {
            "group_id": "my_group_id",
            "isolation_level": "read_commited"
          },
          "key_meta": "Message-Id"
        }
      }
    }
  ]
}
```

## Driver fields

| Field | Description |
|-------|-------------|
| `cluster` | Broker connection (brokers, TLS, SASL) |
| `group` | Consumer group settings (`id` or `group_id`) |
| `key_meta` | HTTP header name for the Kafka message key |

Driver config matches `reader` in `backend/pubsub/subscriber/kafka`, except topics come from `consumer.topic`.

### Cluster fields

| Field | Default | Description |
|-------|---------|-------------|
| `brokers` | — | Kafka broker addresses (required) |
| `client_id` | Pucora version string | Client identifier |
| `client_tls` | — | mTLS settings |
| `sasl` | — | `{ mechanism, user, password, azure_event_hub, auth_identity }` |
| `metadata_retry_max` | `3` | Reconnect attempts |
| `metadata_retry_backoff` | `250ms` | Reconnect delay |

### Consumer group fields

| Field | Default | Description |
|-------|---------|-------------|
| `group_id` / `id` | agent name | Consumer group ID |
| `isolation_level` | `read_commited` | `read_commited` or `read_uncommited` |
| `session_timeout` | `10s` | Failure detection timeout |
| `heartbeat_interval` | `3s` | Coordinator heartbeat interval |

## Offset commits

Partition offsets are committed after each message is read, even when the backend pipeline fails. Design downstream services for idempotency.

## Example

```bash
make pubsub-async-kafka-compose-test
```

See [`examples/pubsub/async-kafka/`](../examples/pubsub/async-kafka/).

## Related

- [Kafka Advanced Pub/Sub](pubsub-kafka-advanced.md)
- [Basic Pub/Sub](pubsub.md)
