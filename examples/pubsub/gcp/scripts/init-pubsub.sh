#!/bin/sh
set -eu

PROJECT="${PUBSUB_PROJECT_ID:-pucora-smoke}"
EMULATOR_HOST="${PUBSUB_EMULATOR_HOST:-pubsub:8085}"
BASE_URL="http://${EMULATOR_HOST}/v1/projects/${PROJECT}"

#region agent log
agent_log() {
  hypothesis_id="$1"
  location="$2"
  message="$3"
  data="${4:-{}}"
  payload="{\"sessionId\":\"575372\",\"hypothesisId\":\"${hypothesis_id}\",\"location\":\"${location}\",\"message\":\"${message}\",\"data\":${data},\"timestamp\":$(date +%s000)}"
  echo "$payload" >&2
  if [ -n "${DEBUG_LOG_FILE:-}" ]; then
    printf '%s\n' "$payload" >> "$DEBUG_LOG_FILE" || true
  fi
}
#endregion

#region agent log
agent_log "H1" "init-pubsub.sh:start" "starting pubsub emulator init" \
  "{\"project\":\"${PROJECT}\",\"emulatorHost\":\"${EMULATOR_HOST}\"}"
#endregion

topic_ready=0
i=1
while [ "$i" -le 30 ]; do
  http_code="$(curl -s -o /tmp/topic-body.txt -w "%{http_code}" -X PUT "${BASE_URL}/topics/events" || true)"
  #region agent log
  agent_log "H2" "init-pubsub.sh:topic-attempt" "topic create attempt" \
    "{\"attempt\":${i},\"httpCode\":\"${http_code}\"}"
  #endregion
  if [ "$http_code" = "200" ] || [ "$http_code" = "409" ]; then
    topic_ready=1
    break
  fi
  sleep 2
  i=$((i + 1))
done

if [ "$topic_ready" -ne 1 ]; then
  body="$(tr -d '\n' < /tmp/topic-body.txt | head -c 200)"
  #region agent log
  agent_log "H2" "init-pubsub.sh:topic-fail" "emulator not ready for topic create" \
    "{\"attempts\":30,\"lastBody\":\"${body}\"}"
  #endregion
  exit 1
fi

sub_payload="{\"topic\":\"projects/${PROJECT}/topics/events\"}"
http_code="$(curl -s -o /tmp/sub-body.txt -w "%{http_code}" \
  -H 'content-type: application/json' \
  -X PUT \
  -d "$sub_payload" \
  "${BASE_URL}/subscriptions/events-sub" || true)"
body="$(tr -d '\n' < /tmp/sub-body.txt | head -c 200)"
#region agent log
agent_log "H4" "init-pubsub.sh:subscription" "subscription create result" \
  "{\"httpCode\":\"${http_code}\",\"body\":\"${body}\"}"
#endregion

if [ "$http_code" != "200" ] && [ "$http_code" != "409" ]; then
  exit 1
fi

#region agent log
agent_log "H1" "init-pubsub.sh:done" "GCP Pub/Sub emulator initialized" "{}"
#endregion
echo "GCP Pub/Sub emulator initialized"
