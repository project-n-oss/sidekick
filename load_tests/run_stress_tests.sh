#!/bin/bash

run_and_log_metrics() {
    local COMMAND=$1
    local OUTPUT_FILE=$2

    # Get number of CPUs
    local NUM_CPUS=$(lscpu -p | grep -v '^#' | wc -l)
    local CPU_HEADERS=""
    for ((i=0;i<NUM_CPUS;i++)); do 
        CPU_HEADERS+="cpu$i,"
    done 

    echo "timestamp, memory, total_cpu, $CPU_HEADERS" > ${OUTPUT_FILE}

    local START_TIME=$(date +%s)
    (while true; do 
      # First call to mpstat to start the timer
      mpstat -P ALL >/dev/null
      local ELAPSED_TIME=$(( $(date +%s) - START_TIME ))
      echo \
        ${ELAPSED_TIME}, \
        $(free -m | awk 'NR==2{printf "%.2f", $3*100/$2}'), \
        $(mpstat -P ALL 1 1 | awk '/Average/' | awk 'NR>=2 && $12 ~ /[0-9.]+/ {printf "%.2f,", 100 - $12}') \
        >> ${OUTPUT_FILE}
    done) &

    local METRICS_PID=$!

    ${COMMAND}

    kill ${METRICS_PID}
}

npm run bundle

run_and_log_metrics "npm run get-original-stress" "get-original-stress-resources.csv"
run_and_log_metrics "npm run get-sidekick-stress" "get-sidekick-stress-resources.csv"

./upload_results.sh