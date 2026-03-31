export type InitTakeoutUploadResponse = {
	upload_id: string;
	chunk_size: number;
};

export type CompleteTakeoutUploadResponse = {
	job_id: number;
	message: string;
};
