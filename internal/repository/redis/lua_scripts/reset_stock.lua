-- KEYS[1]: Key kho hàng
-- ARGV[1]: Số lượng set mới

redis.call('set',KEYS[1],ARGV[1])
return 1