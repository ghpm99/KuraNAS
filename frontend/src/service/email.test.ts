jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		post: jest.fn(),
		put: jest.fn(),
		delete: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	createGoogleAuthUrl,
	deleteEmailAccount,
	getEmailAccounts,
	getEmailAiProvider,
	getMicrosoftDeviceCodeStatus,
	startMicrosoftDeviceCode,
	updateEmailAccountSyncEnabled,
	updateEmailAiProvider,
} from './email';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
	put: jest.Mock;
	delete: jest.Mock;
};

const sampleAccount = {
	id: 1,
	provider: 'google',
	address: 'owner@gmail.com',
	display_name: 'Owner',
	status: 'linked',
	sync_enabled: true,
	last_sync_at: null,
	last_error: '',
	created_at: '2026-06-12T10:00:00Z',
};

describe('service/email', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('lists e-mail accounts', async () => {
		mockedApi.get.mockResolvedValue({ data: [sampleAccount] });

		const result = await getEmailAccounts();

		expect(mockedApi.get).toHaveBeenCalledWith('/email/accounts');
		expect(result).toEqual([sampleAccount]);
	});

	it('deletes an account and returns the backend message', async () => {
		mockedApi.delete.mockResolvedValue({ data: { message: 'removida' } });

		const message = await deleteEmailAccount(1);

		expect(mockedApi.delete).toHaveBeenCalledWith('/email/accounts/1');
		expect(message).toBe('removida');
	});

	it('updates sync enabled', async () => {
		mockedApi.put.mockResolvedValue({});

		await updateEmailAccountSyncEnabled(2, false);

		expect(mockedApi.put).toHaveBeenCalledWith('/email/accounts/2/sync-enabled', {
			sync_enabled: false,
		});
	});

	it('creates the google auth url', async () => {
		mockedApi.post.mockResolvedValue({ data: { auth_url: 'https://accounts.google.com/x' } });

		const result = await createGoogleAuthUrl();

		expect(mockedApi.post).toHaveBeenCalledWith('/email/accounts/google/auth-url');
		expect(result.auth_url).toBe('https://accounts.google.com/x');
	});

	it('starts the microsoft device code flow', async () => {
		const dto = {
			user_code: 'ABC123',
			verification_uri: 'https://microsoft.com/devicelogin',
			expires_in: 900,
			message: 'abra e digite',
		};
		mockedApi.post.mockResolvedValue({ data: dto });

		const result = await startMicrosoftDeviceCode();

		expect(mockedApi.post).toHaveBeenCalledWith('/email/accounts/microsoft/device-code');
		expect(result).toEqual(dto);
	});

	it('reads the microsoft device code status', async () => {
		mockedApi.get.mockResolvedValue({ data: { status: 'pending' } });

		const result = await getMicrosoftDeviceCodeStatus();

		expect(mockedApi.get).toHaveBeenCalledWith('/email/accounts/microsoft/device-code/status');
		expect(result.status).toBe('pending');
	});

	it('reads the e-mail AI provider preference', async () => {
		mockedApi.get.mockResolvedValue({ data: { provider: 'ollama' } });

		const result = await getEmailAiProvider();

		expect(mockedApi.get).toHaveBeenCalledWith('/email/settings/provider');
		expect(result.provider).toBe('ollama');
	});

	it('updates the e-mail AI provider preference', async () => {
		mockedApi.put.mockResolvedValue({ data: { provider: 'anthropic' } });

		const result = await updateEmailAiProvider('anthropic');

		expect(mockedApi.put).toHaveBeenCalledWith('/email/settings/provider', {
			provider: 'anthropic',
		});
		expect(result.provider).toBe('anthropic');
	});
});
