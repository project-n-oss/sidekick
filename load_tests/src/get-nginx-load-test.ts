import http from 'k6/http';
import { check } from 'k6';
// @ts-ignore
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.1.0/index.js';
import { getLoadTestOptions } from './options';
import { SignedRequest } from './signed-request';

// docker run --env-file ./nginx_settings.env --publish 80:80  ghcr.io/nginxinc/nginx-s3-gateway/nginx-oss-s3-gateway:latest

// This test will get the original file through sidekick
const signedRequest = SignedRequest({
    bucket: 'project-n-sidekick-router-test',
    key: '100GB/parquet/zstd/call_center/part-00000-tid-3044319830967113622-f94b7d07-f853-4fc5-900c-c62b22276c2e-1169-1.c000.zstd.parquet',
    endpoint: 'http://localhost:80',
});

export const options = getLoadTestOptions;

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
        'get-nginx-load-test-summary.json': JSON.stringify(data, null, 2),
    };
}
