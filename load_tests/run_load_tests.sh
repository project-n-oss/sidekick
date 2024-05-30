#!/bin/bash

npm run bundle
npm run get-original-load
npm run get-sidekick-load
npm run get-sidekick-crunched-load
npm run get-nginx-load
./upload_results.sh