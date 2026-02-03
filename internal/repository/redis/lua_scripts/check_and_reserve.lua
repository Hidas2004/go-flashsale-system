-- KEYS[1]: Key kho hàng (product:stock:uuid)
-- KEYS[2]: Key giữ chỗ (product:reserved:uuid)
-- ARGV[1]: Số lượng mua
-- ARGV[2]: UserID
-- ARGV[3]: TTL (Time To Live - giây)

local stock_key = KEYS[1]
local reserved_key = KEYS[2]
local quantity = tonumber(ARGV[1])
local user_id = ARGV[2]
local ttl = tonumber(ARGV[3])

--1 check kho
local current_stock = tonumber(redis.call('get',stock_key) or '0')
if current_stock < quantity then
    return 0
end

--2 trừ kho
redis.call('decrby',stock_key,quantity)

--3 giữ chỗ
local reservation_id = user_id .. ":" .. tostring(os.time())
redis.call('hset',reserved_key,reservation_id,quantity)

--4 đặt TTL cho key giữ chỗ
redis.call('expire',reserved_key,ttl)
return reservation_id
