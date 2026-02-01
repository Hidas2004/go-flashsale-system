-- KEY[1] : key chưa tồn kho của san phẩm
-- Key[2] : key chưa lịch sữ mua hàng của user
-- ARGV[1] : số lượng khách muốn mua(quantity)
-- ARGV[2] :giới hạn mua tối đa của mỗi user(limit)
-- ARGV[3] : UserID để chwck limit
--Hash Map  ()

local stock_key = KEYS[1]
local bought_key = KEYS[2]
local buy_qty = tonumber(ARGV[1])
local limit_qty = tonumber(ARGV[2])
local user_id = ARGV[3]

-- 1 Kiểm tra tồn kho
local current_stock = tonumber(redis.call('get',stock_key) or "0")
if current_stock < buy_qty then
    return -1
end

-- 2 Kiểm tra giới hạn mua hàng của user
local user_bought = tonumber(redis.call('hget',bought_key,user_id) or "0")
-- 'hget' là lệnh lấy giá trị trong Hash Map  
--  Xem user này (user_id) đã mua bao nhiêu cái trong key lịch sử (bought_key)
if user_bought + buy_qty > limit_qty then
    return -2
end

-- 3 thực hiện trừ kho
redis.call('decrby',stock_key,buy_qty)
--Cộng thêm số lượng vừa mua buy_qty vào lịch sử của user_id trong Hash bought_key.
redis.call('hincrby',bought_key,user_id,buy_qty)

return 1
