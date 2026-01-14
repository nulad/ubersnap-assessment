# Database Schema

This directory contains the PostgreSQL database schema for the Ubersnap coupon system.

## Files

- `schema.sql` - Main database schema with tables, constraints, indexes, and triggers
- `test_schema.sql` - Test script to validate all constraints and business rules

## Schema Overview

### Tables

#### `coupons`
Stores coupon information with amount tracking.

Columns:
- `id` - Serial primary key
- `name` - Unique coupon identifier (VARCHAR 255)
- `amount` - Total initial amount (non-negative integer)
- `remaining_amount` - Amount available for claims (non-negative integer)
- `created_at` - Timestamp when coupon was created
- `updated_at` - Timestamp when coupon was last modified

#### `claims`
Tracks user claims with unique constraint to prevent duplicates.

Columns:
- `id` - Serial primary key
- `user_id` - User identifier (VARCHAR 255)
- `coupon_name` - Reference to the coupon claimed
- `claimed_at` - Timestamp when claim was made

## Constraints

1. **Unique Constraints**:
   - `coupons.name` - Prevents duplicate coupon names
   - `(user_id, coupon_name)` in claims - Prevents duplicate claims by same user for same coupon

2. **Foreign Key**:
   - `claims.coupon_name` references `coupons.name` with CASCADE delete

3. **Check Constraints**:
   - `amount >= 0` - Prevents negative amounts
   - `remaining_amount >= 0` - Prevents negative remaining amounts
   - `remaining_amount <= amount` - Ensures remaining never exceeds total

## Indexes

- `idx_coupons_name` - On coupons.name for faster lookups
- `idx_claims_user_id` - On claims.user_id for user queries
- `idx_claims_coupon_name` - On claims.coupon_name for coupon queries
- `idx_claims_user_claimed_at` - Composite index for user claim history

## Setup

1. Install PostgreSQL
2. Create database:
   ```sql
   CREATE DATABASE ubersnap;
   ```
3. Apply schema:
   ```bash
   psql -d ubersnap -f schema.sql
   ```
4. Run tests (optional):
   ```bash
   psql -d ubersnap -f test_schema.sql
   ```

## Business Rules Enforced

- Users can only claim a specific coupon once
- Coupon amounts cannot be negative
- Remaining amount cannot exceed total amount
- Deleting a coupon removes all associated claims
- All timestamps are automatically managed
