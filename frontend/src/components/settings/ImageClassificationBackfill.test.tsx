import { fireEvent, render, screen } from '@testing-library/react';
import ImageClassificationBackfill from './ImageClassificationBackfill';

const mockStartBackfill = jest.fn();
const mockUseHook = jest.fn();

jest.mock('./useImageClassificationBackfill', () => ({
    __esModule: true,
    default: () => mockUseHook(),
}));

const createState = (overrides: Record<string, unknown> = {}) => ({
    t: (key: string, vars?: Record<string, unknown>) =>
        vars ? `${key}:${JSON.stringify(vars)}` : key,
    pendingCount: 4,
    isLoading: false,
    hasError: false,
    isStarting: false,
    startBackfill: mockStartBackfill,
    ...overrides,
});

describe('components/settings/ImageClassificationBackfill', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseHook.mockReturnValue(createState());
    });

    it('renders without crashing and shows the pending count', () => {
        render(<ImageClassificationBackfill />);
        expect(
            screen.getByText(/IMAGE_CLASSIFY_BACKFILL_PENDING_LABEL/)
        ).toBeInTheDocument();
    });

    it('shows a loading indicator while counting', () => {
        mockUseHook.mockReturnValue(createState({ isLoading: true }));
        render(<ImageClassificationBackfill />);
        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('shows an error alert when the count fails', () => {
        mockUseHook.mockReturnValue(createState({ hasError: true }));
        render(<ImageClassificationBackfill />);
        expect(
            screen.getByText('IMAGE_CLASSIFY_BACKFILL_TOAST_ERROR')
        ).toBeInTheDocument();
    });

    it('shows the empty label and disables the button when nothing is pending', () => {
        mockUseHook.mockReturnValue(createState({ pendingCount: 0 }));
        render(<ImageClassificationBackfill />);
        expect(screen.getByText('IMAGE_CLASSIFY_BACKFILL_NONE')).toBeInTheDocument();
        expect(
            screen.getByRole('button', { name: 'IMAGE_CLASSIFY_BACKFILL_BUTTON' })
        ).toBeDisabled();
    });

    it('triggers the backfill on click', () => {
        render(<ImageClassificationBackfill />);
        const button = screen.getByRole('button', {
            name: 'IMAGE_CLASSIFY_BACKFILL_BUTTON',
        });
        expect(button).not.toBeDisabled();
        fireEvent.click(button);
        expect(mockStartBackfill).toHaveBeenCalled();
    });

    it('disables the button when the disabled prop is set', () => {
        render(<ImageClassificationBackfill disabled />);
        expect(
            screen.getByRole('button', { name: 'IMAGE_CLASSIFY_BACKFILL_BUTTON' })
        ).toBeDisabled();
    });
});
