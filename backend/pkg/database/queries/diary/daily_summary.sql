WITH
    DAY AS (
        SELECT
            $1::date AS ref_date,
            ($1::date + INTERVAL '1 day') AS next_date
    ),
    activities AS (
        SELECT
            id,
            NAME,
            description,
            GREATEST(start_time, DAY.ref_date) AS start_time,
            LEAST(COALESCE(end_time, DAY.next_date), DAY.next_date) AS end_time
        FROM
            activity_diary,
            DAY
        WHERE
            start_time < DAY.next_date
            AND (
                end_time >= DAY.ref_date
                OR end_time IS NULL
            )
    )
SELECT
    DAY.ref_date AS date,
    COUNT(*) AS total_activities,
    SUM(
        EXTRACT(
            EPOCH
            FROM
                end_time - start_time
        )
    )::INT AS total_time_spent_seconds,
    (
        SELECT
            ROW_TO_JSON(a)
        FROM
            (
                SELECT
                    NAME,
                    EXTRACT(
                        EPOCH
                        FROM
                            end_time - start_time
                    )::INT AS duration_seconds,
                    TO_CHAR((end_time - start_time), 'HH24:MI:SS') AS duration_formatted
                FROM
                    activities
                ORDER BY
                    duration_seconds DESC
                LIMIT
                    1
            ) a
    ) AS longest_activity
FROM
    activities,
    DAY
GROUP BY
    DAY.ref_date;