export type OllamaModel = {
	name: string;
	size: number;
	digest: string;
	modified_at: string;
	family?: string;
	parameter_size?: string;
	quantization_level?: string;
};

export type OllamaStatus = {
	reachable: boolean;
	version?: string;
	base_url: string;
	models: OllamaModel[];
};

export type PullModelResponse = {
	job_id: number;
};
