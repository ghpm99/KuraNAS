import { render } from '@testing-library/react';
import TakeoutImportScreen from './TakeoutImportScreen';
import type { TakeoutUploadState } from './useTakeoutUpload';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

let currentUploadState: TakeoutUploadState = 'idle';

jest.mock('./useTakeoutUpload', () => ({
	__esModule: true,
	default: () => ({
		state: currentUploadState,
		progress: 0,
		fileName: '',
		jobId: null,
		importId: null,
		errorMessage: '',
		progressMessage: '',
		selectFile: jest.fn(),
		startUpload: jest.fn(),
		reset: jest.fn(),
	}),
}));

jest.mock('./TakeoutDropZone', () => ({
	__esModule: true,
	default: () => <div />,
}));

const dispatchBeforeUnload = (): BeforeUnloadEvent => {
	const event = new Event('beforeunload', { cancelable: true }) as BeforeUnloadEvent;
	window.dispatchEvent(event);
	return event;
};

describe('TakeoutImportScreen leave guard', () => {
	it.each<TakeoutUploadState>(['uploading', 'completing'])(
		'asks for confirmation before leaving while %s',
		(uploadState) => {
			currentUploadState = uploadState;
			render(<TakeoutImportScreen />);
			expect(dispatchBeforeUnload().defaultPrevented).toBe(true);
		},
	);

	it.each<TakeoutUploadState>(['idle', 'selecting', 'done', 'error'])(
		'does not block leaving while %s',
		(uploadState) => {
			currentUploadState = uploadState;
			render(<TakeoutImportScreen />);
			expect(dispatchBeforeUnload().defaultPrevented).toBe(false);
		},
	);

	it('removes the guard when the screen unmounts', () => {
		currentUploadState = 'uploading';
		const { unmount } = render(<TakeoutImportScreen />);
		unmount();
		expect(dispatchBeforeUnload().defaultPrevented).toBe(false);
	});
});
