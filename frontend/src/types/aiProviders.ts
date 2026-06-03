export type AIProviderName = 'ollama' | 'openai' | 'anthropic';

export type AIProviderParams = {
	keep_alive?: string;
	timeout_seconds?: number;
	max_retries?: number;
	retry_backoff_ms?: number;
};

export type AIProviderDto = {
	name: AIProviderName;
	enabled: boolean;
	model: string;
	base_url: string;
	priority: number;
	params: AIProviderParams;
	requires_api_key: boolean;
	api_key_configured: boolean;
};

export type UpdateAIProviderRequest = {
	enabled: boolean;
	model: string;
	base_url: string;
	priority: number;
	params: AIProviderParams;
};
