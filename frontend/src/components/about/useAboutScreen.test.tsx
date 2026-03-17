import { act, renderHook } from '@testing-library/react';
import { useAboutScreen } from './useAboutScreen';
const mockClipboard = {
	writeText: jest.fn(),
};

const useAboutMock = jest.fn(() => ({
	version: '1.0.0',
	path: '/root',
	enable_workers: true,
	uptime: '1h',
	statup_time: 'invalid-date',
	commit_hash: 'abc123',
	platform: 'linux',
	lang: 'en',
	gin_mode: 'release',
	gin_version: '1.0',
	go_version: '1.20',
	node_version: '18',
}));

jest.mock('@/components/providers/aboutProvider/AboutContext', () => ({
	useAbout: () => useAboutMock(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => key,
	}),
}));

describe('useAboutScreen', () => {
	beforeEach(() => {
		jest.useFakeTimers();
		mockClipboard.writeText.mockReset();
		Object.defineProperty(navigator, 'clipboard', {
			value: mockClipboard,
			configurable: true,
		});
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	it('formats runtime details and handles copy feedback', async () => {
		const { result } = renderHook(() => useAboutScreen());

		expect(result.current.runtimeDetails[4].value).toBe('invalid-date');

		mockClipboard.writeText.mockResolvedValue(undefined);
		await act(async () => {
			await result.current.copyCommitHash();
		});
		expect(result.current.copied).toBe(true);

		act(() => {
			jest.runOnlyPendingTimers();
		});
		expect(result.current.copied).toBe(false);
	});

	it('does not copy when commit hash is missing', async () => {
		useAboutMock.mockReturnValueOnce({
			version: '1.0.0',
			path: '/root',
			enable_workers: true,
			uptime: '1h',
			statup_time: '2023-01-01T00:00:00Z',
			commit_hash: '',
			platform: 'linux',
			lang: 'en',
			gin_mode: 'release',
			gin_version: '1.0',
			go_version: '1.20',
			node_version: '18',
		});

		const { result } = renderHook(() => useAboutScreen());

		await act(async () => {
			await result.current.copyCommitHash();
		});

		expect(mockClipboard.writeText).not.toHaveBeenCalled();
		expect(result.current.copied).toBe(false);
	});
});
