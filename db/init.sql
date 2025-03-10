-- Database initialization script for Payment Gateway Integration System

-- Create tables if they don't exist
CREATE TABLE IF NOT EXISTS countries (
                                         id SERIAL PRIMARY KEY,
                                         name VARCHAR(255) NOT NULL UNIQUE,
    code CHAR(2) NOT NULL UNIQUE,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS gateways (
                                        id SERIAL PRIMARY KEY,
                                        name VARCHAR(255) NOT NULL UNIQUE,
    data_format_supported VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS gateway_countries (
                                                 gateway_id INT NOT NULL,
                                                 country_id INT NOT NULL,
                                                 priority INT NOT NULL DEFAULT 0,
                                                 PRIMARY KEY (gateway_id, country_id),
    FOREIGN KEY (gateway_id) REFERENCES gateways(id),
    FOREIGN KEY (country_id) REFERENCES countries(id)
    );

CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL DEFAULT 'password',
    country_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (country_id) REFERENCES countries(id)
    );

CREATE TABLE IF NOT EXISTS transactions (
                                            id SERIAL PRIMARY KEY,
                                            amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    reference_id VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    gateway_id INT NOT NULL,
    country_id INT NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY (gateway_id) REFERENCES gateways(id),
    FOREIGN KEY (country_id) REFERENCES countries(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
    );

-- Insert sample data only if tables are empty
DO $$
BEGIN
    -- Insert countries
    IF NOT EXISTS (SELECT 1 FROM countries LIMIT 1) THEN
        INSERT INTO countries (name, code, currency) VALUES
        ('United States', 'US', 'USD'),
        ('United Kingdom', 'GB', 'GBP'),
        ('Germany', 'DE', 'EUR'),
        ('Japan', 'JP', 'JPY');
END IF;

    -- Insert gateways
    IF NOT EXISTS (SELECT 1 FROM gateways LIMIT 1) THEN
        INSERT INTO gateways (name, data_format_supported) VALUES
        ('PayPal', 'application/json'),
        ('Stripe', 'application/json'),
        ('Adyen', 'application/xml');
END IF;

    -- Insert gateway-country relationships with priorities
    IF NOT EXISTS (SELECT 1 FROM gateway_countries LIMIT 1) THEN
        -- For US (1)
        INSERT INTO gateway_countries (gateway_id, country_id, priority) VALUES
        (1, 1, 1), -- PayPal primary for US
        (2, 1, 2), -- Stripe secondary for US
        (3, 1, 3); -- Adyen tertiary for US

        -- For UK (2)
INSERT INTO gateway_countries (gateway_id, country_id, priority) VALUES
                                                                     (2, 2, 1), -- Stripe primary for UK
                                                                     (1, 2, 2), -- PayPal secondary for UK
                                                                     (3, 2, 3); -- Adyen tertiary for UK

-- For Germany (3)
INSERT INTO gateway_countries (gateway_id, country_id, priority) VALUES
                                                                     (3, 3, 1), -- Adyen primary for Germany
                                                                     (2, 3, 2), -- Stripe secondary for Germany
                                                                     (1, 3, 3); -- PayPal tertiary for Germany

-- For Japan (4)
INSERT INTO gateway_countries (gateway_id, country_id, priority) VALUES
                                                                     (1, 4, 1), -- PayPal primary for Japan
                                                                     (2, 4, 2); -- Stripe secondary for Japan
END IF;

    -- Insert sample users
    IF NOT EXISTS (SELECT 1 FROM users LIMIT 1) THEN
        INSERT INTO users (username, email, country_id) VALUES
        ('user1', 'user1@example.com', 1), -- US user
        ('user2', 'user2@example.com', 2), -- UK user
        ('user3', 'user3@example.com', 3), -- German user
        ('user4', 'user4@example.com', 4); -- Japanese user
END IF;
END $$;