ORDER BY
    (
        SELECT
            MAX(
                CASE
                    WHEN vm.duration ~ '^[0-9]+(\.[0-9]+)?$' THEN vm.duration::numeric
                END
            )
        FROM
            video_metadata vm
        WHERE
            vm.file_id = hf.id
    ) DESC NULLS LAST,
    hf.id DESC
