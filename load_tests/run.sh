#!/bin/bash

npm run bundle
npm run get-original
npm run get-sidekick
npm run get-sidekick-crunched
./upload_results.sh