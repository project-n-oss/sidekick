import crypto from 'k6/crypto';
import http from "k6/http";
import { check } from "k6";
import { AWSConfig, SignatureV4 } from "https://jslib.k6.io/aws/0.7.1/aws.js";

export const options = {
  scenarios: {
    rps_batched: {
      executor: 'constant-arrival-rate',
      duration: '1m',
      rate: 250,
      timeUnit: '1s',
      preAllocatedVUs: 2000 // above 200, start getting TCP read errors (connection reset by peer)
    },
    // ramping_batched: {
    //   executor: 'ramping-vus',
    //   stages: [
    //     { duration: "30s", target: 500 },  // ramp up to 2000 users over 5 seconds
    //     { duration: "10s", target: 500 }, // stay at 2000 users for 10 seconds
    //     { duration: "5s", target: 0 },    // ramp down to 0 users over 5 seconds
    //   ]
    // }
  }
};

const n_unique_files = 10000
const batch_size = 20;

const { AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION, BUCKET } = process.env;

const awsConfig = new AWSConfig({
  region: AWS_REGION,
  accessKeyId: AWS_ACCESS_KEY_ID,
  secretAccessKey: AWS_SECRET_ACCESS_KEY,
});

const bucket = BUCKET;
const objSize = 6144;

const signer = new SignatureV4({
  service: "s3",
  region: awsConfig.region,
  credentials: {
    accessKeyId: awsConfig.accessKeyId,
    secretAccessKey: awsConfig.secretAccessKey,
  },
});

// Function to generate random alphanumeric characters
function generateRandomString(length) {
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
  let result = '';
  const charactersLength = characters.length;
  for (let i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
  }
  return result;
};

// Generate a unique filename using a timestamp and a random string
function generateUniqueFilename() {
  const timestamp = Date.now();
  const randomString = generateRandomString(8);
  const filename = `${timestamp}-${randomString}.txt`;
  return filename;
};

function getRandomItems(arr, n) {
  const selectedIndices = new Set();
  while (selectedIndices.size < n) {
    const randomIndex = Math.floor(Math.random() * arr.length);
    selectedIndices.add(randomIndex);
  }
  const randomItems = Array.from(selectedIndices).map((index) => arr[index]);
  return randomItems;
}

export function setup() {
  console.log("Generating " + n_unique_files + " random files.")
  const files = [];
  for (let i = 0; i < n_unique_files; i++) {
    const uniqueFilename = generateUniqueFilename();
    const signedRequest = signer.sign(
      {
        method: "PUT",
        protocol: "http",
        hostname: "localhost:7075",
        path: `/${bucket}/${uniqueFilename}`,
        headers: {},
        uriEscapePath: false,
        applyChecksum: true,
      },
      {
        signingDate: new Date(),
        signingService: "s3",
      }
    );

    files.push({
      method: "PUT",
      url: signedRequest.url,
      body: crypto.randomBytes(objSize),
      params: {
        headers: signedRequest.headers,
      },
    });
  }
  return { files: files };
}

export default async function (data) {
  const files = data.files
  const batch = getRandomItems(files, batch_size)

  const responses = http.batch(batch);
  // check that respones returned 200
  responses.forEach((res) => {
    check(res, {
      "is status 200": (r) => r.status === 200,
    });
  });
}
