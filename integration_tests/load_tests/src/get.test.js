import http from "k6/http";
import { check } from "k6";
import { AWSConfig, SignatureV4 } from "https://jslib.k6.io/aws/0.7.1/aws.js";

export const options = {
  stages: [
    { duration: "5s", target: 100 },
    { duration: "10s", target: 100 },
    { duration: "5s", target: 0 },
  ],
};

const { AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION, BUCKET } =
  process.env;

const awsConfig = new AWSConfig({
  region: AWS_REGION,
  accessKeyId: AWS_ACCESS_KEY_ID,
  secretAccessKey: AWS_SECRET_ACCESS_KEY,
});

const bucket = BUCKET;
const key = "animals/1.csv";

export default async function () {
  const signer = new SignatureV4({
    service: "s3",
    region: awsConfig.region,
    credentials: {
      accessKeyId: awsConfig.accessKeyId,
      secretAccessKey: awsConfig.secretAccessKey,
    },
  });

  const signedRequest = signer.sign(
    {
      method: "GET",
      protocol: "http",
      hostname: "localhost:7075",
      path: `/${bucket}/${key}`,
      headers: {},
      uriEscapePath: false,
      applyChecksum: true,
    },
    {
      signingDate: new Date(),
      signingService: "s3",
    }
  );

  const res = http.get(signedRequest.url, { headers: signedRequest.headers });
  check(res, {
    "is status 200": (r) => r.status === 200,
    "contains data": (r) => r.body.length > 0,
  });
}
