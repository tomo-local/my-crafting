#!/usr/bin/env zsh
# Usage: ./bench.zsh [requests]
REQUESTS=${1:-100}
PROXY="http://localhost:8080/"

run_bench() {
  local label=$1
  local start end elapsed
  start=$(( EPOCHREALTIME * 1000 ))
  for i in $(seq 1 $REQUESTS); do
    curl -s "$PROXY" > /dev/null
  done
  end=$(( EPOCHREALTIME * 1000 ))
  elapsed=$(( int(end - start) ))
  echo "$label: ${elapsed}ms (${REQUESTS} reqs, avg $(( elapsed / REQUESTS ))ms/req)"
}

echo "=== Connection Pool Benchmark ($REQUESTS requests) ==="

# upstream 起動
(cd go && go run upstream/main.go -port 9001 -id upstream-1 2>/dev/null) &
UPSTREAM_PID=$!
sleep 1

echo ""
echo ">>> pool-size 0 (no pooling)"
(cd go && go run main.go -upstreams localhost:9001 -port 8080 -pool-size 0 2>/dev/null) &
PROXY_PID=$!
sleep 0.5
run_bench "no-pool"
kill $PROXY_PID 2>/dev/null
sleep 0.3

echo ""
echo ">>> pool-size 10"
(cd go && go run main.go -upstreams localhost:9001 -port 8080 -pool-size 10 2>/dev/null) &
PROXY_PID=$!
sleep 0.5
run_bench "pool-10"
kill $PROXY_PID 2>/dev/null

kill $UPSTREAM_PID 2>/dev/null
echo ""
echo "=== Done ==="
