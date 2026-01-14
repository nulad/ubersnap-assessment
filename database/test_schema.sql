-- Test script to validate database schema constraints
-- Run this after creating the schema to verify all constraints work

-- Test 1: Insert valid coupon
INSERT INTO coupons (name, amount, remaining_amount) 
VALUES ('test-coupon', 100, 100);

-- Test 2: Try to insert coupon with negative amount (should fail)
-- INSERT INTO coupons (name, amount, remaining_amount) 
-- VALUES ('bad-coupon', -10, -10);

-- Test 3: Try to insert coupon with remaining_amount > amount (should fail)
-- INSERT INTO coupons (name, amount, remaining_amount) 
-- VALUES ('bad-coupon-2', 50, 60);

-- Test 4: Try to insert duplicate coupon name (should fail)
-- INSERT INTO coupons (name, amount, remaining_amount) 
-- VALUES ('test-coupon', 50, 50);

-- Test 5: Create valid claim
INSERT INTO claims (user_id, coupon_name) 
VALUES ('user-123', 'test-coupon');

-- Test 6: Try to create duplicate claim (should fail)
-- INSERT INTO claims (user_id, coupon_name) 
-- VALUES ('user-123', 'test-coupon');

-- Test 7: Try to claim non-existent coupon (should fail)
-- INSERT INTO claims (user_id, coupon_name) 
-- VALUES ('user-456', 'non-existent');

-- Test 8: Verify foreign key cascade - delete coupon should delete claims
DELETE FROM coupons WHERE name = 'test-coupon';

-- Verify claims are deleted
SELECT COUNT(*) FROM claims WHERE coupon_name = 'test-coupon';

-- Test 9: Insert multiple coupons for comprehensive testing
INSERT INTO coupons (name, amount, remaining_amount) VALUES 
('coupon-1', 100, 100),
('coupon-2', 200, 200),
('coupon-3', 300, 300);

-- Test 10: Multiple users claiming different coupons
INSERT INTO claims (user_id, coupon_name) VALUES 
('user-1', 'coupon-1'),
('user-2', 'coupon-1'),
('user-1', 'coupon-2'),
('user-3', 'coupon-3');

-- Test 11: Try to claim same coupon twice (should fail)
-- INSERT INTO claims (user_id, coupon_name) 
-- VALUES ('user-1', 'coupon-1');

-- View final state
SELECT 'Coupons:' as table_name;
SELECT * FROM coupons ORDER BY name;

SELECT 'Claims:' as table_name;
SELECT * FROM claims ORDER BY user_id, coupon_name;
