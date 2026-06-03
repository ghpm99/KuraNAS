jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		post: jest.fn(),
		delete: jest.fn(),
	},
}));

import { apiBase } from './index';
import { deleteOllamaModel, getOllamaStatus, listOllamaModels, pullOllamaModel } from './ollama';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
	delete: jest.Mock;
};

describe('service/ollama', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('gets status', async () => {
		const response = { reachable: true, version: '0.5.0', base_url: 'http://x', models: [] };
		mockedApi.get.mockResolvedValue({ data: response });

		const result = await getOllamaStatus();

		expect(mockedApi.get).toHaveBeenCalledWith('/ai/ollama/status');
		expect(result).toEqual(response);
	});

	it('lists models', async () => {
		mockedApi.get.mockResolvedValue({ data: [] });
		await listOllamaModels();
		expect(mockedApi.get).toHaveBeenCalledWith('/ai/ollama/models');
	});

	it('pulls a model', async () => {
		mockedApi.post.mockResolvedValue({ data: { job_id: 7 } });

		const result = await pullOllamaModel('llama3.1');

		expect(mockedApi.post).toHaveBeenCalledWith('/ai/ollama/models/pull', { model: 'llama3.1' });
		expect(result).toEqual({ job_id: 7 });
	});

	it('deletes a model', async () => {
		mockedApi.delete.mockResolvedValue({ data: {} });
		await deleteOllamaModel('llama3.1:latest');
		expect(mockedApi.delete).toHaveBeenCalledWith('/ai/ollama/models/llama3.1%3Alatest');
	});
});
