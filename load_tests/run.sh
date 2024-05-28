#!/bin/bash

npm run bundle
npm run get-original
npm run get-sidekick
npm run get-nginx-test
npm run get-sidekick-crunched
./upload_results.sh