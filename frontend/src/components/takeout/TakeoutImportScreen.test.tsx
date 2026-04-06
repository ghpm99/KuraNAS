import { render, screen, fireEvent } from '@testing-library/react';
import TakeoutImportScreen from './TakeoutImportScreen';
import type { TakeoutUploadState } from './useTakeoutUpload';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockSelectFile = jest.fn();
const mockStartUpload = jest.fn();
const mockReset = jest.fn();

let hookState: {
	state: TakeoutUploadState;
	progress: number;
	fileName: string;
	jobId: number | null;
	errorMessage: string;
	progressMessage: string;
};

jest.mock('./useTakeoutUpload', () => ({
	__esModule: true,
	default: () => ({
		...hookState,
		selectFile: mockSelectFile,
		startUpload: mockStartUpload,
		reset: mockReset,
	}),
}));

jest.mock('./TakeoutDropZone', () => ({
	__esModule: true,
	default: ({ onSelectFile }: { onSelectFile: (f: File) => void }) => (
		<div data-testid="drop-zone" onClick={() => onSelectFile(new File([], 'test.zip'))} />
	),
}));

describe('TakeoutImportScreen', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		hookState = {
			state: 'idle',
			progress: 0,
			fileName: '',
			jobId: null,
			errorMessage: '',
			progressMessage: '',
		};
	});

	it('renders title and description', () => {
		render(<TakeoutImportScreen />);
		expect(screen.getByText('TAKEOUT_PAGE_TITLE')).toBeInTheDocument();
		expect(screen.getByText('TAKEOUT_PAGE_DESCRIPTION')).toBeInTheDocument();
	});

	it('renders drop zone in idle state', () => {
		render(<TakeoutImportScreen />);
		expect(screen.getByTestId('drop-zone')).toBeInTheDocument();
	});

	it('shows file name when a file is selected', () => {
		hookState.fileName = 'takeout.zip';
		hookState.state = 'selecting';
		render(<TakeoutImportScreen />);
		expect(screen.getByText('takeout.zip')).toBeInTheDocument();
	});

	it('shows progress during upload', () => {
		hookState.state = 'uploading';
		hookState.progress = 50;
		hookState.progressMessage = '50 MB / 100 MB';
		render(<TakeoutImportScreen />);
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.getByText('50 MB / 100 MB')).toBeInTheDocument();
	});

	it('shows processing text during completing state', () => {
		hookState.state = 'completing';
		hookState.progress = 100;
		render(<TakeoutImportScreen />);
		expect(screen.getByText('TAKEOUT_PROCESSING')).toBeInTheDocument();
	});

	it('shows success message on complete', () => {
		hookState.state = 'done';
		hookState.jobId = 42;
		render(<TakeoutImportScreen />);
		expect(screen.getByText(/TAKEOUT_UPLOAD_COMPLETE/)).toBeInTheDocument();
		expect(screen.getByText(/job #42/)).toBeInTheDocument();
	});

	it('shows success without job id', () => {
		hookState.state = 'done';
		hookState.jobId = null;
		render(<TakeoutImportScreen />);
		expect(screen.getByText(/TAKEOUT_UPLOAD_COMPLETE/)).toBeInTheDocument();
	});

	it('shows error message on failure', () => {
		hookState.state = 'error';
		hookState.errorMessage = 'Something went wrong';
		render(<TakeoutImportScreen />);
		expect(screen.getByText('Something went wrong')).toBeInTheDocument();
	});

	it('disables upload button in idle state', () => {
		render(<TakeoutImportScreen />);
		const button = screen.getByText('TAKEOUT_UPLOADING');
		expect(button.closest('button')).toBeDisabled();
	});

	it('calls reset on reset button click', () => {
		render(<TakeoutImportScreen />);
		fireEvent.click(screen.getByText('SETTINGS_RESET'));
		expect(mockReset).toHaveBeenCalled();
	});
});
