jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		put: jest.fn(),
	},
}));

import { apiBase } from './index';
import { getAIProviders, updateAIProvider } from './aiProviders';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	put: jest.Mock;
};

describe('service/aiProviders', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('gets providers', async () => {
		const response = [{ name: 'ollama', enabled: true, model: 'llama3.1' }];
		mockedApi.get.mockResolvedValue({ data: response });

		const result = await getAIProviders();

		expect(mockedApi.get).toHaveBeenCalledWith('/ai/providers');
		expect(result).toEqual(response);
	});

	it('updates a provider', async () => {
		const request = {
			enabled: true,
			model: 'qwen2.5',
			base_url: 'http://localhost:11434',
			priority: 0,
			params: {},
		};
		const response = { name: 'ollama', ...request };
		mockedApi.put.mockResolvedValue({ data: response });

		const result = await updateAIProvider('ollama', request);

		expect(mockedApi.put).toHaveBeenCalledWith('/ai/providers/ollama', request);
		expect(result).toEqual(response);
	});
});
