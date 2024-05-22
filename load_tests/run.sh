#!/bin/bash

npm run bundle
npm run get-original
npm run get-sidekick-router
npm run get-sidekick-router-crunched
./upload_results.sh