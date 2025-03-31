CREATE TABLE IF NOT EXISTS calendars (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    refresh_token VARCHAR(255),
    created_at TIMESTAMP(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at TIMESTAMP(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
);

CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(255) NOT NULL,
    calendar_id VARCHAR(255),
    summary VARCHAR(255) NOT NULL,
    start TIMESTAMP,
    end TIMESTAMP,
    status VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at TIMESTAMP(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    FOREIGN KEY (calendar_id) REFERENCES calendars(id)
);

CREATE TABLE IF NOT EXISTS channel_histories (
    calendar_id VARCHAR(255),
    start_time TIMESTAMP(3) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    expiration TIMESTAMP(3) NOT NULL,
    is_stopped BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at TIMESTAMP(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (calendar_id, start_time),
    FOREIGN KEY (calendar_id) REFERENCES calendars(id),
    INDEX idx_calendar_expiration (calendar_id, expiration)
);

CREATE TABLE IF NOT EXISTS sync_histories (
    calendar_id VARCHAR(255),
    sync_time TIMESTAMP(3) NOT NULL,
    next_sync_token VARCHAR(255) NOT NULL,
    updated_event_count INT NOT NULL,
    created_at TIMESTAMP(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at TIMESTAMP(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (calendar_id, sync_time),
    FOREIGN KEY (calendar_id) REFERENCES calendars(id)
);
