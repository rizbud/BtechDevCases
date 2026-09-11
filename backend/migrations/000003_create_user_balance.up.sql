CREATE VIEW user_balances AS
SELECT 
    u.id AS user_id,
    COALESCE(SUM(t.net_amount), 0.00) AS balance
FROM users u
LEFT JOIN (
    -- Outgoing transactions subtract from balance
    SELECT from_user_id AS user_id, -amount AS net_amount FROM transactions
    UNION ALL
    -- Incoming transactions add to balance
    SELECT to_user_id AS user_id, amount AS net_amount FROM transactions
) t ON u.id = t.user_id
GROUP BY u.id;