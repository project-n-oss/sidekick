import http from "k6/http";
import { check } from "k6";
import { AWSConfig, SignatureV4 } from "https://jslib.k6.io/aws/0.7.1/aws.js";

export const options = {
  stages: [
    { duration: "20s", target: 2000 },  // ramp up to 2000 users over 5 seconds
    { duration: "10s", target: 2000 }, // stay at 2000 users for 10 seconds
    { duration: "5s", target: 0 },    // ramp down to 0 users over 5 seconds
  ],
};

const batch_size = 25;

const { AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION, BUCKET } = process.env;

const awsConfig = new AWSConfig({
  region: AWS_REGION,
  accessKeyId: AWS_ACCESS_KEY_ID,
  secretAccessKey: AWS_SECRET_ACCESS_KEY,
});

const bucket = BUCKET;
const objSize = 6144;

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
  const randomString = generateRandomString(8); // Adjust the length as needed
  const filename = `${timestamp}-${randomString}.txt`; // You can change the file extension as needed
  return filename;
};

export function setup() {
  const signer = new SignatureV4({
    service: "s3",
    region: awsConfig.region,
    credentials: {
      accessKeyId: awsConfig.accessKeyId,
      secretAccessKey: awsConfig.secretAccessKey,
    },
  });
  const batch = [];
  for (let i = 0; i < batch_size; i++) {
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

    batch.push({
      method: "PUT",
      url: signedRequest.url,
      body: generateRandomString(objSize),
      params: {
        headers: signedRequest.headers,
      },
    });
  }
  return { data: batch };
}

export default async function (data) {
  // Make the batch request
  const responses = http.batch(data.data);

  // check that respones returned 200
  responses.forEach((res) => {
    check(res, {
      "is status 200": (r) => r.status === 200,
    });
  });
}
