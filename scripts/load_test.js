import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

// Load the queries from test_queries.json
const queries = new SharedArray('queries', function () {
    return JSON.parse(open('../test_queries.json'));
});

export const options = {
    stages: [
        { duration: '30s', target: 20 }, // ramp up to 20 users over 30 seconds
        { duration: '1m', target: 20 },  // stay at 20 users for 1 minute
        { duration: '30s', target: 0 },  // ramp down to 0 users
    ],
    thresholds: {
        http_req_duration: ['p(95)<2000'], // 95% of requests must complete below 2s
    },
};

export default function () {
    // Pick a random query from the array
    const query = queries[Math.floor(Math.random() * queries.length)];

    // Construct the URL with the query parameter
    // We use encodeURIComponent to ensure special characters are handled correctly
    const url = `http://localhost:8000/query?query=${encodeURIComponent(query)}`;

    const res = http.get(url);

    // Validate the response
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response is not empty': (r) => r.body.length > 0,
    });

    // Wait for a short period between requests
    sleep(1);
}
