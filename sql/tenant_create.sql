CREATE TABLE IF NOT EXISTS "api_keys" (
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    token VARCHAR(255) NOT NULL
);
--
--
CREATE TABLE IF NOT EXISTS "scores" (
    score INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- 
-- Generate random numbers
WITH RECURSIVE seq(n, score, ts) AS (
    SELECT 0,
        ABS(RANDOM() % 191) + 10,
        -- random number between 10 and 200
        DATETIME('now')
    UNION ALL
    SELECT n + 1,
        ABS(RANDOM() % 191) + 10,
        DATETIME(ts, '+5 minutes')
    FROM seq
    WHERE n < 54000
)
INSERT INTO scores (score, created_at)
SELECT score,
    ts
FROM seq;