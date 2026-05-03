import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { Counter, Rate } from 'k6/metrics';

// Custom metrics to track benchmark accuracy
const accuracyRate = new Rate('accuracy_rate');
const cacheHitRate = new Rate('actual_cache_hit_rate');
const unexpectedMisses = new Counter('unexpected_misses');

// Load metadata which contains both queries and expected outcomes
const metadata = new SharedArray('metadata', function () {
    const data = JSON.parse(open('../test_metadata.json'));
    return data.queries;
});

export const options = {
    scenarios: {
        benchmark: {
            executor: 'per-vu-iterations',
            vus: 1, // Start with 1 VU to strictly follow the discovery sequence in metadata
            iterations: metadata.length,
            maxDuration: '2h',
        },
    },
    thresholds: {
        accuracy_rate: ['rate>0.95'], // Expect high alignment with metadata
        http_req_duration: ['p(95)<60000'], // Relaxed to 60s for local execution
    },
};

export default function (data) {
    // Get the current iteration index to pick the right query from the sequence
    const index = __ITER;
    if (index >= metadata.length) return;

    const item = metadata[index];
    const url = `http://localhost:8000/query?query=${encodeURIComponent(item.query)}`;

    const res = http.get(url, { timeout: '300s' });

    const actualCache = res.headers['X-Cache'] || 'MISS';
    const isCorrect = actualCache === item.expected;

    // Record metrics
    accuracyRate.add(isCorrect);
    cacheHitRate.add(actualCache === 'HIT');
    
    if (actualCache === 'MISS' && item.expected === 'HIT') {
        unexpectedMisses.add(1);
    }

    // Standard K6 checks
    check(res, {
        'status is 200': (r) => r.status === 200,
        'X-Cache header present': (r) => r.headers['X-Cache'] !== undefined,
        'matches metadata expectation': () => isCorrect,
    });
}
