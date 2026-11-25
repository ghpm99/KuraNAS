export type InitTakeoutUploadResponse = {
	upload_id: string;
	chunk_size: number;
	received_size: number;
	resumed: boolean;
};

export type CompleteTakeoutUploadResponse = {
	job_id: number;
	import_id: number;
	message: string;
};

export type TakeoutUploadStatus = {
	upload_id: string;
	size: number;
	received_size: number;
};

export type TakeoutImport = {
	id: number;
	job_id: number | null;
	file_name: string;
	size: number;
	state: string;
	total_entries: number;
	processed_entries: number;
	imported: number;
	already_existing: number;
	ignored: number;
	failed_entries: number;
	error: string;
	created_at: string;
	ended_at: string | null;
};

export type ListTakeoutImportsParams = {
	page?: number;
	pageSize?: number;
};
