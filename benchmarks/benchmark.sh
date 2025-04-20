#!/bin/bash

# Benchmark config
TIME=1      # duration of each test (seconds)
CONNECTIONS=5  # concurrent users
PORTS=(8001 8002 8003 8004 8005 8007)
NAMES=("FastAPI" "Starlette" "Sanic" "Tornado" "Flask" "FastPaze")
FILES=("fastapi_app.py" "starlette_app.py" "sanic_app.py" "tornado_app.py" "flask_app.py" "fastpaze_app.py")

echo "🔬 Benchmarking APIs for $DURATION seconds with $CONNECTIONS connections..."

for i in "${!NAMES[@]}"; do
  NAME=${NAMES[$i]}
  FILE=${FILES[$i]}
  PORT=${PORTS[$i]}

  echo
  echo "🚀 Running $NAME on port $PORT..."

  if [[ "$NAME" == "FastAPI" ]]; then
    uvicorn fastapi_app:app --port=$PORT > /dev/null 2>&1 &
  elif [[ "$NAME" == "Flask" ]]; then
    python3 "$FILE" > /dev/null 2>&1 &
  elif [[ "$NAME" == "Starlette" ]]; then
    uvicorn starlette_app:app --port=$PORT > /dev/null 2>&1 &
  elif [[ "$NAME" == "FastPaze" ]]; then
    python3 "$FILE" > /dev/null 2>&1 &
  elif [[ "$NAME" == "Sanic" ]]; then
    python3 "$FILE" > /dev/null 2>&1 &
  elif [[ "$NAME" == "Tornado" ]]; then
    python3 "$FILE" > /dev/null 2>&1 &       
  fi

  PID=$!
  sleep 10

  echo "⚡ Benchmarking $NAME..."
  autocannon -c $CONNECTIONS -d $TIME http://localhost:$PORT/hello

  echo "🛑 Stopping $NAME..."
  kill $PID
  wait $PID 2>/dev/null
done

echo
echo "✅ All benchmarks complete!"
