#!/bin/bash

BUCKET="s3://project-n-sidekick-router-test"

# Get the current date
YEAR=$(date +%Y)
MONTH=$(date +%m)
DAY=$(date +%d)


# Filename for storing the current index
INDEX_FILE="index.txt"

DATE_PATH="${BUCKET}/results/${YEAR}-${MONTH}-${DAY}"

# Get the current index from the S3 bucket
aws s3 cp "${DATE_PATH}/${INDEX_FILE}" .
INDEX=$(cat ${INDEX_FILE})

# If index file was not present or empty, initialize the index to 1
if [ -z "$INDEX" ]
then
  INDEX=0
fi

((INDEX++))

BUCKET_PATH="${DATE_PATH}/${INDEX}"

# Search for files and upload to S3 bucket
find -type f -name "*-summary.json" | while read -r file
do
  filename=$(basename "$file")
  echo "Uploading ${filename} to ${BUCKET_PATH}"
  aws s3 cp "${file}" "${BUCKET_PATH}/${filename}"
done

# Increment the index and store in the bucket for next time
echo $INDEX > $INDEX_FILE
aws s3 cp $INDEX_FILE ${DATE_PATH}/${INDEX_FILE}
rm ${INDEX_FILE}