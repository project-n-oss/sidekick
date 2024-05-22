import http from 'k6/http';
import { check } from 'k6';
// @ts-ignore
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.1.0/index.js';
import { getOptions } from './get-options';
import { SignedRequest } from './signed-request';

// This test will use sidekick to try and read a deleted and crunched file
const signedRequest = SignedRequest({
    bucket: 'project-n-sidekick-router-test',
    key: 'crunched/100GB/parquet/zstd/call_center/part-00000-tid-3044319830967113622-f94b7d07-f853-4fc5-900c-c62b22276c2e-1169-1.c000.zstd.parquet',
    endpoint: 'http://localhost:7075',
});

export const options = getOptions;

export default async function () {
    const res = http.get(signedRequest.url, { headers: signedRequest.headers });
    check(res, {
        'is status 500': (r) => r.status === 500,
        'contains data': (r) => r.body !== undefined,
    });
}

export function handleSummary(data: any) {
    return {
        stdout: textSummary(data, { indent: ' ', enableColors: true }),
        'get-sidekick-router-crunched-summary.json': JSON.stringify(data, null, 2),
    };
}
