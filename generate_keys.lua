for i = 1, 10000, 1 do
    redis.call("SELECT", 1)
    redis.call("SET", "ZzZ_MYKEY_ZzZ_"..i.."key", i)
end

return "Ok!"
