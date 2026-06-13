import { apiBase } from '@/service';
import type {
	EmailAccountDto,
	EmailAiProvider,
	EmailDeviceCodeDto,
	EmailDeviceCodeStatusDto,
	EmailProviderPreferenceDto,
	GoogleAuthUrlDto,
} from '@/types/email';

export const getEmailAccounts = async (): Promise<EmailAccountDto[]> => {
	const response = await apiBase.get<EmailAccountDto[]>('/email/accounts');
	return response.data;
};

export const deleteEmailAccount = async (id: number): Promise<string | undefined> => {
	const response = await apiBase.delete<{ message?: string }>(`/email/accounts/${id}`);
	return response.data?.message;
};

export const updateEmailAccountSyncEnabled = async (
	id: number,
	syncEnabled: boolean
): Promise<void> => {
	await apiBase.put(`/email/accounts/${id}/sync-enabled`, { sync_enabled: syncEnabled });
};

export const createGoogleAuthUrl = async (): Promise<GoogleAuthUrlDto> => {
	const response = await apiBase.post<GoogleAuthUrlDto>('/email/accounts/google/auth-url');
	return response.data;
};

export const startMicrosoftDeviceCode = async (): Promise<EmailDeviceCodeDto> => {
	const response = await apiBase.post<EmailDeviceCodeDto>(
		'/email/accounts/microsoft/device-code'
	);
	return response.data;
};

export const getMicrosoftDeviceCodeStatus = async (): Promise<EmailDeviceCodeStatusDto> => {
	const response = await apiBase.get<EmailDeviceCodeStatusDto>(
		'/email/accounts/microsoft/device-code/status'
	);
	return response.data;
};

export const getEmailAiProvider = async (): Promise<EmailProviderPreferenceDto> => {
	const response = await apiBase.get<EmailProviderPreferenceDto>('/email/settings/provider');
	return response.data;
};

export const updateEmailAiProvider = async (
	provider: EmailAiProvider
): Promise<EmailProviderPreferenceDto> => {
	const response = await apiBase.put<EmailProviderPreferenceDto>('/email/settings/provider', {
		provider,
	});
	return response.data;
};
