export type StorageRootDto = {
	id: number;
	path: string;
	label: string;
	enabled: boolean;
	created_at: string;
};

export type CreateStorageRootRequest = {
	path: string;
	label?: string;
	enabled?: boolean;
};

export type UpdateStorageRootRequest = {
	label?: string;
	enabled?: boolean;
};
