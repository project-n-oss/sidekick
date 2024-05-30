export const getLoadTestOptions = {
    stages: [
        { duration: '5s', target: 100 },
        { duration: '30s', target: 100 },
        { duration: '5s', target: 0 },
    ],
};

export const getStressTestOptions = {
    thresholds: {
        'http_req_duration{scenario:vu1}': [],
        'http_req_duration{scenario:vu5}': [],
        'http_req_duration{scenario:vu10}': [],
        'http_req_duration{scenario:vu20}': [],
        'http_req_duration{scenario:vu40}': [],
        'http_req_duration{scenario:vu80}': [],
        'http_req_duration{scenario:vu160}': [],
    },
    scenarios: {
        vu1: {
            executor: 'constant-vus',
            vus: 1,
            startTime: '0s',
            duration: '10s',
        },
        vu5: {
            executor: 'constant-vus',
            vus: 5,
            startTime: '10s',
            duration: '10s',
        },
        vu10: {
            executor: 'constant-vus',
            vus: 10,
            startTime: '20s',
            duration: '10s',
        },
        vu20: {
            executor: 'constant-vus',
            vus: 20,
            startTime: '30s',
            duration: '10s',
        },
        vu40: {
            executor: 'constant-vus',
            vus: 40,
            startTime: '40s',
            duration: '10s',
        },
        vu80: {
            executor: 'constant-vus',
            vus: 80,
            startTime: '50s',
            duration: '10s',
        },
        vu160: {
            executor: 'constant-vus',
            vus: 160,
            startTime: '60s',
            duration: '10s',
        },
    },
};

export const getSmokeTestOptions = {
    iterations: 1000,
};
