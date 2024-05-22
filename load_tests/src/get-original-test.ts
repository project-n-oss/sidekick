import http from 'k6/http';
import { check } from 'k6';
// @ts-ignore
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.1.0/index.js';
import { getOptions } from './get-options';
import { SignedRequest } from './signed-request';

// this test calls s3 directly
const signedRequest = SignedRequest({
    bucket: 'project-n-sidekick-router-test',
    key: '100GB/parquet/zstd/call_center/part-00000-tid-3044319830967113622-f94b7d07-f853-4fc5-900c-c62b22276c2e-1169-1.c000.zstd.parquet',
    endpoint: 'https://s3.us-east-1.amazonaws.com',
});

export const options = getOptions;

export default async function () {
    const res = http.get(signedRequest.url, { headers: signedRequest.headers });
    check(res, {
        'is status 200': (r) => r.status === 200,
        'contains data': (r) => r.body !== undefined,
    });
}

export function handleSummary(data: any) {
    return {
        stdout: textSummary(data, { indent: ' ', enableColors: true }),
        'get-original-summary.json': JSON.stringify(data, null, 2),
    };
}
