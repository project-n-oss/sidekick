// @ts-ignore
import { AWSConfig, SignatureV4, Endpoint } from 'https://jslib.k6.io/aws/0.12.0/signature.js';

interface SignedRequestProps {
    bucket: string;
    key: string;
    endpoint: string;
}

interface SignedRequestResult {
    url: string;
    headers: Record<string, string>;
}

export function SignedRequest(props: SignedRequestProps): SignedRequestResult {
    const { AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN, AWS_REGION } = process.env;
    const awsConfig = new AWSConfig({
        region: AWS_REGION,
        accessKeyId: AWS_ACCESS_KEY_ID,
        secretAccessKey: AWS_SECRET_ACCESS_KEY,
        sessionToken: AWS_SESSION_TOKEN,
    });

    const signer = new SignatureV4({
        service: 's3',
        region: awsConfig.region,
        uriEscapePath: false,
        applyChecksum: true,
        credentials: {
            accessKeyId: awsConfig.accessKeyId,
            secretAccessKey: awsConfig.secretAccessKey,
            sessionToken: awsConfig.sessionToken,
        },
    });

    const endpoint = new Endpoint(props.endpoint);

    const signedRequest = signer.sign(
        {
            method: 'GET',
            endpoint: endpoint,
            headers: {},
            path: `/${props.bucket}/${props.key}`,
        },
        {
            signingDate: new Date(),
            signingService: 's3',
        },
    );
    return signedRequest;
}
