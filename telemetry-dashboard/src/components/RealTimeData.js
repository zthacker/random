import React, { useEffect, useState } from 'react';
import { toast, ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

const TelemetryTable = () => {
    const [telemetryData, setTelemetryData] = useState([]);

    useEffect(() => {
        // Open WebSocket connection
        const ws = new WebSocket('ws://localhost:4000/api/v1/telemetry/ws');

        ws.onmessage = (event) => {
            try {
                // Parse the incoming message
                const data = JSON.parse(event.data);

                // Update the table with the new telemetry data
                setTelemetryData((prevData) => [data, ...prevData].slice(0, 10)); // Keep only the last 10 entries

                // Check for anomalies and trigger a toast notification
                if (data.anomalies && data.anomalies.length > 0) {
                    toast.error(`Anomaly detected: ${data.anomalies.join(', ')}`, {
                        position: "top-right",
                        autoClose: 5000,
                    });
                }
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };

        // Cleanup on component unmount
        return () => ws.close();
    }, []);

    return (
        <div>
            <h1>Recent Telemetry Data</h1>
            <table border="1" style={{ width: '100%', textAlign: 'left' }}>
                <thead>
                <tr>
                    <th>ID</th>
                    <th>Timestamp</th>
                    <th>Packet ID</th>
                    <th>Temperature</th>
                    <th>Battery</th>
                    <th>Altitude</th>
                    <th>Signal</th>
                    <th>Anomalies</th>
                </tr>
                </thead>
                <tbody>
                {telemetryData.map((entry, index) => (
                    <tr key={index}>
                        <td>{entry.id}</td>
                        <td>{new Date(entry.timestamp).toLocaleString()}</td>
                        <td>{entry.packet_id}</td>
                        <td>{entry.temperature.toFixed(2)}</td>
                        <td>{entry.battery.toFixed(2)}</td>
                        <td>{entry.altitude.toFixed(2)}</td>
                        <td>{entry.signal.toFixed(2)}</td>
                        <td>
                            {entry.anomalies && entry.anomalies.length > 0 ? (
                                entry.anomalies.join(', ')
                            ) : (
                                'None'
                            )}
                        </td>
                    </tr>
                ))}
                </tbody>
            </table>
            {/* Toast notification container */}
            <ToastContainer />
        </div>
    );
};

export default TelemetryTable;
