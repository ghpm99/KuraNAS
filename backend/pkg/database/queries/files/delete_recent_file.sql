DELETE FROM recent_file
WHERE
    ip_address = $1
    AND file_id = $2;