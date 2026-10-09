import { act, render, renderHook, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import FileDetailsProvider from './fileDetailsProvider';
import useFileDetails from './useFileDetails';

const mockUseFile = jest.fn();

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

const wrapper = ({ children }: { children: ReactNode }) => (
    <FileDetailsProvider>{children}</FileDetailsProvider>
);

const folder = { id: 5, name: 'photos', type: 1 } as any;

describe('useFileDetails', () => {
    beforeEach(() => mockUseFile.mockReturnValue({ selectedItem: null }));

    it('is unavailable and inert outside the provider', () => {
        const { result } = renderHook(() => useFileDetails());

        expect(result.current.isAvailable).toBe(false);
        expect(result.current.explicitTarget).toBeNull();
        expect(() => act(() => result.current.openDetails(folder))).not.toThrow();
        expect(() => act(() => result.current.closeDetails())).not.toThrow();
    });

    it('opens and closes details for any item', () => {
        const { result } = renderHook(() => useFileDetails(), { wrapper });

        expect(result.current.isAvailable).toBe(true);

        act(() => result.current.openDetails(folder));
        expect(result.current.explicitTarget).toBe(folder);

        act(() => result.current.closeDetails());
        expect(result.current.explicitTarget).toBeNull();
    });

    it('drops the explicit target when the user navigates to another item', () => {
        const { result, rerender } = renderHook(() => useFileDetails(), { wrapper });

        act(() => result.current.openDetails(folder));
        expect(result.current.explicitTarget).toBe(folder);

        mockUseFile.mockReturnValue({ selectedItem: { id: 9 } });
        rerender();

        expect(result.current.explicitTarget).toBeNull();
    });

    it('keeps the target for the item that was being viewed when it was opened', () => {
        mockUseFile.mockReturnValue({ selectedItem: { id: 5 } });
        const { result, rerender } = renderHook(() => useFileDetails(), { wrapper });

        act(() => result.current.openDetails(folder));
        rerender();

        expect(result.current.explicitTarget).toBe(folder);
    });

    it('renders children inside the provider', () => {
        render(
            <FileDetailsProvider>
                <span>child</span>
            </FileDetailsProvider>
        );

        expect(screen.getByText('child')).toBeInTheDocument();
    });
});
