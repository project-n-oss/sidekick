#!/bin/bash

npm run bundle
npm run get-original-stress
npm run get-sidekick-stress
./upload_results.sh