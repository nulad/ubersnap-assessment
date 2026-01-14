-- Ubersnap Database Schema
-- PostgreSQL schema with proper constraints, indexes, and foreign keys

-- Enable UUID extension if needed for future use
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Coupons table
-- Stores coupon information with amount tracking
CREATE TABLE coupons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    amount INTEGER NOT NULL CHECK (amount >= 0),
    remaining_amount INTEGER NOT NULL CHECK (remaining_amount >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Claims table
-- Tracks user claims with unique constraint to prevent duplicates
CREATE TABLE claims (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    coupon_name VARCHAR(255) NOT NULL,
    claimed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Critical: Prevents duplicate claims by same user for same coupon
    CONSTRAINT unique_user_coupon UNIQUE (user_id, coupon_name),
    
    -- Foreign key relationship
    CONSTRAINT fk_coupon 
        FOREIGN KEY (coupon_name) 
        REFERENCES coupons(name) 
        ON DELETE CASCADE
);

-- Indexes for performance
-- Index on coupons.name for faster lookups
CREATE INDEX idx_coupons_name ON coupons(name);

-- Index on claims.user_id for user-specific queries
CREATE INDEX idx_claims_user_id ON claims(user_id);

-- Index on claims.coupon_name for coupon-specific queries
CREATE INDEX idx_claims_coupon_name ON claims(coupon_name);

-- Composite index for common query patterns (user_id, claimed_at)
CREATE INDEX idx_claims_user_claimed_at ON claims(user_id, claimed_at);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_coupons_updated_at 
    BEFORE UPDATE ON coupons 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add check to ensure remaining_amount never exceeds amount
ALTER TABLE coupons ADD CONSTRAINT check_remaining_not_exceed_amount 
    CHECK (remaining_amount <= amount);

-- Comments for documentation
COMMENT ON TABLE coupons IS 'Stores coupon information with amount tracking';
COMMENT ON TABLE claims IS 'Tracks user claims with unique constraint to prevent duplicates';
COMMENT ON COLUMN coupons.amount IS 'Total initial amount of the coupon';
COMMENT ON COLUMN coupons.remaining_amount IS 'Remaining amount available for claims';
COMMENT ON COLUMN claims.user_id IS 'Identifier for the user who claimed the coupon';
COMMENT ON COLUMN claims.coupon_name IS 'Reference to the coupon that was claimed';
