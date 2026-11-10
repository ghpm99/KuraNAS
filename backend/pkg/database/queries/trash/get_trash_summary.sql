-- How much the trash holds: item count and total bytes.
SELECT
    COUNT(*),
    COALESCE(SUM(size), 0)
FROM
    trash_item;
