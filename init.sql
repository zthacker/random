-- Anomaly Flags Mapping:
-- Bit 0: Temperature anomaly
-- Bit 1: Battery anomaly
-- Bit 2: Altitude anomaly
-- Bit 3: Signal anomaly

CREATE TABLE telemetry (
                           id SERIAL PRIMARY KEY,
                           timestamp TIMESTAMPTZ NOT NULL,
                           packet_id INTEGER NOT NULL,
                           seq_flags INTEGER NOT NULL,
                           seq_count INTEGER NOT NULL,
                           subsystem_id INTEGER NOT NULL,
                           temperature REAL NOT NULL,
                           battery REAL NOT NULL,
                           altitude REAL NOT NULL,
                           signal REAL NOT NULL,
                           anomaly_flags INTEGER NOT NULL
);

-- Create the notify function
CREATE OR REPLACE FUNCTION notify_telemetry_update()
    RETURNS TRIGGER AS $$
DECLARE
    inserted_data JSON;
BEGIN
    -- Aggregate all rows inserted in this statement into a JSON array
    SELECT json_agg(row_to_json(t))
    INTO inserted_data
    FROM telemetry AS t
    WHERE t.id IN (SELECT id FROM telemetry WHERE id >= CURRVAL('telemetry_id_seq') - TG_NARGS);

    -- Send the aggregated JSON array as a notification
    IF inserted_data IS NOT NULL THEN
        PERFORM pg_notify('telemetry_update', inserted_data::text);
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER telemetry_update_trigger
    AFTER INSERT ON telemetry
    FOR EACH STATEMENT
EXECUTE FUNCTION notify_telemetry_update();
