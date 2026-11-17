UPDATE takeout_import
SET state = 'running'
WHERE job_id = $1;
