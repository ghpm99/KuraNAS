DELETE FROM recent_file
WHERE
    ip_address = ?
    AND file_id = ?;