import { render, screen, fireEvent } from '@testing-library/react';
import TakeoutImportScreen from './TakeoutImportScreen';
import { createQueueEntry, type TakeoutQueueEntry } from './takeoutUploadQueue';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockAddFile = jest.fn();
const mockStartQueue = jest.fn();
const mockClearSettled = jest.fn();

let hookState: {
    entries: TakeoutQueueEntry[];
    isBusy: boolean;
    hasQueuedEntries: boolean;
};

const entryWith = (changes: Partial<TakeoutQueueEntry>): TakeoutQueueEntry => ({
    ...createQueueEntry(new File(['x'], 'takeout.zip'), 'entry-1'),
    ...changes,
});

jest.mock('./useTakeoutUploadQueue', () => ({
    __esModule: true,
    default: () => ({
        ...hookState,
        addFile: mockAddFile,
        startQueue: mockStartQueue,
        clearSettled: mockClearSettled,
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
        hookState = { entries: [], isBusy: false, hasQueuedEntries: false };
    });

    it('renders title and description', () => {
        render(<TakeoutImportScreen />);
        expect(screen.getByText('TAKEOUT_PAGE_TITLE')).toBeInTheDocument();
        expect(screen.getByText('TAKEOUT_PAGE_DESCRIPTION')).toBeInTheDocument();
    });

    it('renders the Google export guide with a safe external link', () => {
        render(<TakeoutImportScreen />);
        expect(screen.getByText('TAKEOUT_GUIDE_TITLE')).toBeInTheDocument();
        expect(screen.getByText('TAKEOUT_GUIDE_STEP_SELECT')).toBeInTheDocument();
        expect(screen.getByText('TAKEOUT_GUIDE_STEP_FORMAT')).toBeInTheDocument();
        expect(screen.getByText('TAKEOUT_GUIDE_STEP_DROP')).toBeInTheDocument();
        const takeoutLink = screen.getByRole('link', { name: 'TAKEOUT_GUIDE_LINK' });
        expect(takeoutLink).toHaveAttribute('href', 'https://takeout.google.com');
        expect(takeoutLink).toHaveAttribute('target', '_blank');
        expect(takeoutLink).toHaveAttribute('rel', 'noopener noreferrer');
    });

    it('renders drop zone in idle state', () => {
        render(<TakeoutImportScreen />);
        expect(screen.getByTestId('drop-zone')).toBeInTheDocument();
    });

    it('lists every entry of the queue', () => {
        hookState.entries = [
            entryWith({ id: 'a' }),
            entryWith({ id: 'b', file: new File(['x'], 'other.zip') }),
        ];
        render(<TakeoutImportScreen />);
        expect(screen.getByText(/takeout\.zip/)).toBeInTheDocument();
        expect(screen.getByText(/other\.zip/)).toBeInTheDocument();
    });

    it('forwards the selected file to the queue', () => {
        render(<TakeoutImportScreen />);
        fireEvent.click(screen.getByTestId('drop-zone'));
        expect(mockAddFile).toHaveBeenCalledWith(expect.objectContaining({ name: 'test.zip' }));
    });

    it('shows progress for an entry being uploaded', () => {
        hookState.entries = [entryWith({ state: 'uploading', uploadedBytes: 1 })];
        render(<TakeoutImportScreen />);
        expect(screen.getByRole('progressbar')).toBeInTheDocument();
        expect(screen.getByText('TAKEOUT_UPLOAD_PROGRESS')).toBeInTheDocument();
    });

    it('shows the error message of a failed entry', () => {
        hookState.entries = [entryWith({ state: 'error', errorMessage: 'Something went wrong' })];
        render(<TakeoutImportScreen />);
        expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    });

    it('disables the start button without queued entries', () => {
        render(<TakeoutImportScreen />);
        expect(screen.getByText('TAKEOUT_START_UPLOAD').closest('button')).toBeDisabled();
    });

    it('enables the start button with queued entries and starts the queue', () => {
        hookState.entries = [entryWith({})];
        hookState.hasQueuedEntries = true;
        render(<TakeoutImportScreen />);
        const startButton = screen
            .getByText('TAKEOUT_START_UPLOAD')
            .closest('button') as HTMLElement;
        expect(startButton).toBeEnabled();
        fireEvent.click(startButton);
        expect(mockStartQueue).toHaveBeenCalled();
    });

    it('disables the start button while the queue is busy', () => {
        hookState.hasQueuedEntries = true;
        hookState.isBusy = true;
        render(<TakeoutImportScreen />);
        expect(screen.getByText('TAKEOUT_START_UPLOAD').closest('button')).toBeDisabled();
    });

    it('clears settled entries on the clear button click', () => {
        render(<TakeoutImportScreen />);
        fireEvent.click(screen.getByText('TAKEOUT_CLEAR'));
        expect(mockClearSettled).toHaveBeenCalled();
    });
});
