jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		post: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	getActivityDiarySummary,
	getActivityDiaryEntries,
	createActivityDiaryEntry,
	duplicateActivityDiaryEntry,
} from './activityDiary';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
};

describe('service/activityDiary', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('gets activity diary summary', async () => {
		const payload = { totalEntries: 10, streak: 5 };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getActivityDiarySummary();

		expect(mockedApi.get).toHaveBeenCalledWith('/diary/summary');
		expect(result).toEqual(payload);
	});

	it('gets activity diary entries', async () => {
		const payload = { items: [{ id: 1, name: 'Entry 1' }], total: 1 };
		mockedApi.get.mockResolvedValue({ data: payload });

		const result = await getActivityDiaryEntries();

		expect(mockedApi.get).toHaveBeenCalledWith('/diary/');
		expect(result).toEqual(payload);
	});

	it('creates activity diary entry', async () => {
		const entry = { id: 1, name: 'Workout', description: 'Leg day' };
		mockedApi.post.mockResolvedValue({ data: entry });

		const result = await createActivityDiaryEntry({
			name: 'Workout',
			description: 'Leg day',
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/diary/', {
			name: 'Workout',
			description: 'Leg day',
		});
		expect(result).toEqual(entry);
	});

	it('duplicates activity diary entry', async () => {
		const entry = { id: 2, name: 'Workout copy' };
		mockedApi.post.mockResolvedValue({ data: entry });

		const result = await duplicateActivityDiaryEntry(1);

		expect(mockedApi.post).toHaveBeenCalledWith('/diary/copy', { ID: 1 });
		expect(result).toEqual(entry);
	});
});
