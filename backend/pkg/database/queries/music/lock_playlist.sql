SELECT id
FROM playlist
WHERE id = $1
FOR UPDATE;
