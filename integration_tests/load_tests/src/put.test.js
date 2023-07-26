import http from "k6/http";
import { check } from "k6";
import { AWSConfig, SignatureV4 } from "https://jslib.k6.io/aws/0.7.1/aws.js";

export const options = {
  stages: [
    { duration: "5s", target: 100 },  // 100 users for 5 seconds
    { duration: "10s", target: 700},   // ramp up to 700 users over the next 3 minutes
    { duration: "30s", target: 700 },  // stay at 700 users for 2 minutes
    { duration: "10s", target: 10000 }, // ramp up to 1000 users over the next 2 minutes
    { duration: "2m", target: 10000 }, // stay at 1000 users for 2 minutes
    { duration: "30s", target: 50000 },  // ramp up to 50000 users over the next 30 seconds
    { duration: "1m", target: 50000 }, // stay at 50000 users for 1 minute
    { duration: "30s", target: 0 },    // ramp down to 0 users over the next 1 minute
  ],
};

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

export default async function () {
  const signer = new SignatureV4({
    service: "s3",
    region: awsConfig.region,
    credentials: {
      accessKeyId: awsConfig.accessKeyId,
      secretAccessKey: awsConfig.secretAccessKey,
    },
  });

  const uniqueFilename = generateUniqueFilename();
  const someStrData = generateRandomString(objSize);

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

  // Make a PUT request to the signed URL with the file in the body
  const res = http.put(signedRequest.url, someStrData, {
    headers: signedRequest.headers,
  });
  check(res, {
    "is status 200": (r) => r.status === 200,
  });
}
