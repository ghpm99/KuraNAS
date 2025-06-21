SELECT
    hf.id,
    hf.name,
    hf."path",
    hf.parent_path,
    hf.format,
    hf."size",
    hf.updated_at,
    hf.created_at,
    hf.last_interaction,
    hf.last_backup,
    hf."type",
    hf.checksum,
    hf.deleted_at
FROM
    home_file hf
WHERE
    1 = 1
    AND (
        ?
        OR hf.id = ?
    )
    AND (
        ?
        OR hf.name LIKE '%' || ? || '%'
    )
    AND (
        ?
        OR hf."path" = ?
    )
    AND (
        ?
        OR hf."parent_path" = ?
    )
    AND (
        ?
        OR hf.format = ?
    )
    AND (
        ?
        OR hf."type" = ?
    )
    AND (
        ?
        OR hf.deleted_at = ?
    )
    AND case ?
        when 'all' then true
        when 'recent' then hf.id IN (
            SELECT
                file_id
            FROM
                recent_file
        )
        when 'starred' then hf.starred = true
        else true
    end
ORDER BY
    type,
    name,
    - id
LIMIT
    ?
OFFSET
    ?;