export type EmailProvider = 'google' | 'microsoft';

export type EmailAccountStatus = 'linked' | 'error' | 'reauth_required';

export interface EmailAccountDto {
	id: number;
	provider: EmailProvider;
	address: string;
	display_name: string;
	status: EmailAccountStatus;
	sync_enabled: boolean;
	last_sync_at: string | null;
	last_error: string;
	created_at: string;
}

export interface GoogleAuthUrlDto {
	auth_url: string;
}

export interface EmailDeviceCodeDto {
	user_code: string;
	verification_uri: string;
	expires_in: number;
	message: string;
}

export type EmailDeviceCodeStatus = 'idle' | 'pending' | 'linked' | 'expired' | 'error';

export interface EmailDeviceCodeStatusDto {
	status: EmailDeviceCodeStatus;
}

export type EmailAiProvider = 'auto' | 'ollama' | 'openai' | 'anthropic';

export interface EmailProviderPreferenceDto {
	provider: EmailAiProvider;
}
