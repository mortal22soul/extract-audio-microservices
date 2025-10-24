-- Redis Lua script for initializing data structures and helper functions
-- This script sets up the initial Redis data structures for the video converter system

-- Helper function to create a user session
local function create_user_session(user_id, token_hash, expires_at)
    local session_key = "user:session:" .. user_id
    redis.call('HMSET', session_key,
        'token', token_hash,
        'expires', expires_at,
        'lastActivity', redis.call('TIME')[1]
    )
    redis.call('EXPIRE', session_key, 86400) -- 24 hours TTL
    return session_key
end

-- Helper function to update conversion progress
local function update_conversion_progress(job_id, progress, status, estimated_time)
    local progress_key = "conversion:" .. job_id
    redis.call('HMSET', progress_key,
        'progress', progress,
        'status', status,
        'estimatedTime', estimated_time or '',
        'lastUpdate', redis.call('TIME')[1]
    )
    redis.call('EXPIRE', progress_key, 3600) -- 1 hour TTL
    
    -- Publish progress update
    local message = cjson.encode({
        jobId = job_id,
        progress = progress,
        status = status,
        estimatedTime = estimated_time,
        timestamp = redis.call('TIME')[1]
    })
    redis.call('PUBLISH', 'conversion:progress', message)
    
    return progress_key
end

-- Helper function to track user activity
local function track_user_activity(user_id, activity)
    local activity_key = "user:activity:" .. user_id
    local timestamp = redis.call('TIME')[1]
    redis.call('ZADD', activity_key, timestamp, activity)
    
    -- Keep only last 100 activities
    redis.call('ZREMRANGEBYRANK', activity_key, 0, -101)
    redis.call('EXPIRE', activity_key, 604800) -- 7 days TTL
    
    return activity_key
end

-- Helper function for rate limiting
local function check_rate_limit(user_id, endpoint, limit, window)
    local rate_key = "rate_limit:" .. user_id .. ":" .. endpoint
    local current = redis.call('GET', rate_key)
    
    if current == false then
        redis.call('SET', rate_key, 1)
        redis.call('EXPIRE', rate_key, window)
        return { allowed = true, remaining = limit - 1 }
    else
        current = tonumber(current)
        if current >= limit then
            return { allowed = false, remaining = 0 }
        else
            redis.call('INCR', rate_key)
            return { allowed = true, remaining = limit - current - 1 }
        end
    end
end

-- Initialize pub/sub channels
redis.call('PUBLISH', 'system:init', 'Redis initialization started')

-- Create sample data for development
local sample_user_id = "507f1f77bcf86cd799439011"
local sample_token = "sample_jwt_token_hash"
local sample_expires = redis.call('TIME')[1] + 86400

-- Create sample user session
create_user_session(sample_user_id, sample_token, sample_expires)

-- Create sample conversion progress
update_conversion_progress("job_001", 100, "completed", "0")
update_conversion_progress("job_002", 65, "processing", "120")

-- Track sample user activities
track_user_activity(sample_user_id, "uploaded_video")
track_user_activity(sample_user_id, "started_conversion")

-- Set up some cache entries for testing
redis.call('SET', 'config:max_file_size', '104857600') -- 100MB
redis.call('SET', 'config:supported_formats', 'mp4,avi,mov,mkv,webm')
redis.call('SET', 'config:conversion_timeout', '3600') -- 1 hour

-- Create a sorted set for video popularity tracking
redis.call('ZADD', 'video:popularity', 100, '507f1f77bcf86cd799439021')
redis.call('ZADD', 'video:popularity', 85, '507f1f77bcf86cd799439022')
redis.call('ZADD', 'video:popularity', 42, '507f1f77bcf86cd799439023')

redis.call('PUBLISH', 'system:init', 'Redis initialization completed')