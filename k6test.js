import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 50 }, // Ramp-up to 50 users over 30 seconds
        { duration: '1m', target: 50 },  // Stay at 50 users for 1 minute
        { duration: '30s', target: 0 },  // Ramp-down to 0 users over 30 seconds
    ],
};

const endpoints = [
    'http://turionbackend:4000/api/v1/telemetry?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z',
    'http://turionbackend:4000/api/v1/telemetry/current',
    'http://turionbackend:4000/api/v1/anomalies?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=altitude&aggregation=avg',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=battery&aggregation=avg',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=signal&aggregation=avg',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=temperature&aggregation=avg',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=altitude&aggregation=min',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=battery&aggregation=min',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=signal&aggregation=min',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=temperature&aggregation=min',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=altitude&aggregation=max',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=battery&aggregation=max',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=signal&aggregation=max',
    'http://turionbackend:4000/api/v1/telemetry/aggregations?start_time=2024-12-05T01:00:00Z&end_time=2024-12-05T23:00:00Z&metric=temperature&aggregation=max',
];

export default function () {
    // Pick a random endpoint
    const url = endpoints[Math.floor(Math.random() * endpoints.length)];

    // Make the HTTP request
    const res = http.get(url);

    // Check the response status
    check(res, { 'status is 200': (r) => r.status === 200 });

    // Pause between requests
    sleep(1);
}
