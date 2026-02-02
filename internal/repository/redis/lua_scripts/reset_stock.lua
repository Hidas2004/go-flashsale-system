--KEY[1] kho hàng
--KEY[2] sổ giữ chổ
--ARGV[1] sô lương mua
--ARGV[2] user id
--ARGV[3] thời gian giữ chổ

local stock_key = KEYS[1]
local reserve_key = KEYS[2]
local quantity = tonumber(ARGV[1])
local user_id = ARGV[2]
local reserve_time = tonumber(ARGV[3])

--1 kiểm tra tồn kho
local current_stock = tonumber(redis.call('get',stock_key) or "0")
if current_stock < quantity then
    return -1
end

--2 trừ kho 
redis.call('decrby',stock_key,quantity)

-- 3. Tạo một "Phiếu giữ chỗ" tạm thời
-- Key này sẽ tự hủy sau TTL (giây)
local reservation_id = user_id .. ":" .. tostring(os.time())
local member_key = reservation_id

-- Lưu vào Hash Map giữ chỗ: Field=ReservationID, Value=Quantity
redis.call('hset',reserve_key,member_key,quantity)

return reservation_id