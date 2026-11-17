INSERT INTO takeout_import (job_id, file_name, size, state)
VALUES ($1, $2, $3, 'queued')
RETURNING id;
