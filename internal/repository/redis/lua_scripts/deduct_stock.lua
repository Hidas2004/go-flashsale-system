-- KEYS[1]: Key chứa tồn kho sản phẩm (Ví dụ: product:101:stock)
-- KEYS[2]: Key chứa lịch sử mua hàng (Ví dụ: product:101:bought_history - Dạng Hash)
-- ARGV[1]: Số lượng muốn mua (quantity)
-- ARGV[2]: Giới hạn tối đa cho mỗi user (limit)
-- ARGV[3]: User ID người mua

local stockKey = KEYS[1]
local historyKey = KEYS[2]
local quantity = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local userId = ARGV[3]

-- 1. Lấy số lượng user này đã mua trước đó (Nếu chưa mua thì là 0)
local currentBought = tonumber(redis.call('hget', historyKey, userId) or '0')

-- 2. Kiểm tra giới hạn User
if (currentBought + quantity) > limit then
    return -2
end

-- 3. Kiểm tra tồn kho
local currentStock = tonumber(redis.call("GET", stockKey) or "0")
if currentStock < quantity then
    return -1
end

-- 4. Nếu mọi thứ OK -> Thực hiện trừ kho và ghi nhận lịch sử
redis.call("DECRBY", stockKey, quantity)
redis.call("HINCRBY", historyKey, userId, quantity)

return 1
