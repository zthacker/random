import React, { useEffect, useState } from 'react';
import { Line } from 'react-chartjs-2';
import DatePicker from 'react-datepicker';
import 'react-datepicker/dist/react-datepicker.css';
import ErrorBoundary from "./ErrorBoundary";

import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
} from 'chart.js';

// Register Chart.js components
ChartJS.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend
);

const HistoricalGraph = () => {
    const [historicalData, setHistoricalData] = useState([]);
    const [startDate, setStartDate] = useState(new Date(new Date().setDate(new Date().getDate() - 7))); // Default to 7 days ago
    const [endDate, setEndDate] = useState(new Date()); // Default to today
    const [loading, setLoading] = useState(false);

    // Fetch data when startDate or endDate changes
    useEffect(() => {
        const fetchData = async () => {
            setLoading(true);
            const startISO = startDate.toISOString();
            const endISO = endDate.toISOString();
            try {
                const response = await fetch(`http://localhost:4000/api/v1/telemetry?start_time=${startISO}&end_time=${endISO}`);
                if (!response.ok) {
                    throw new Error('Failed to fetch telemetry data');
                }
                const jsonResponse = await response.json();

                // Extract the data field
                if (Array.isArray(jsonResponse.data)) {
                    setHistoricalData(jsonResponse.data);
                } else {
                    console.error('Unexpected response format:', jsonResponse);
                    setHistoricalData([]); // Default to an empty array
                }
            } catch (error) {
                console.error('Error fetching historical data:', error);
                setHistoricalData([]); // Default to an empty array
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, [startDate, endDate]);

    // Prepare data for the chart
    const chartData = {
        labels: historicalData.map((entry) => new Date(entry.timestamp).toLocaleString()),
        datasets: [
            {
                label: 'Temperature',
                data: historicalData.map((entry) => entry.temperature),
                borderColor: 'rgba(75, 192, 192, 1)',
                borderWidth: 2,
            },
            {
                label: 'Battery',
                data: historicalData.map((entry) => entry.battery),
                borderColor: 'rgba(255, 99, 132, 1)',
                borderWidth: 2,
            },
            {
                label: 'Altitude',
                data: historicalData.map((entry) => entry.altitude),
                borderColor: 'rgb(130,116,255)',
                borderWidth: 2,
            },
            {
                label: 'Signal',
                data: historicalData.map((entry) => entry.signal),
                borderColor: 'rgb(251,213,122)',
                borderWidth: 2,
            },
        ],
    };

    return (
        <div>
            <h1>Historical Telemetry Data</h1>
            <div style={{ marginBottom: '20px' }}>
                <label>
                    Start Date & Time:
                    <DatePicker
                        selected={startDate}
                        onChange={(date) => setStartDate(date)}
                        showTimeSelect
                        timeFormat="HH:mm"
                        timeIntervals={15}
                        dateFormat="yyyy-MM-dd HH:mm"
                        maxDate={new Date()}
                    />
                </label>
                <label style={{ marginLeft: '20px' }}>
                    End Date & Time:
                    <DatePicker
                        selected={endDate}
                        onChange={(date) => setEndDate(date)}
                        showTimeSelect
                        timeFormat="HH:mm"
                        timeIntervals={15}
                        dateFormat="yyyy-MM-dd HH:mm"
                        maxDate={new Date()}
                        minDate={startDate}
                    />
                </label>
            </div>
            {loading ? (
                <p>Loading historical data...</p>
            ) : (
                <Line data={chartData} />
            )}
            <div>
                <h2>Recent Anomalies (Most Recent 10)</h2>
                <ul>
                    {historicalData
                        .filter((entry) => entry.anomalies && entry.anomalies.length > 0)
                        .slice(-10) // Display the last 10 anomalies
                        .map((entry, index) => (
                            <li key={index}>
                                {new Date(entry.timestamp).toLocaleString()}: {entry.anomalies.join(', ')}
                            </li>
                        ))}
                </ul>
            </div>
        </div>
    );
};

export default function HistoricalGraphWithBoundary() {
    return (
        <ErrorBoundary>
            <HistoricalGraph />
        </ErrorBoundary>
    );
}
