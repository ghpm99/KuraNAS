import { fireEvent, render, screen } from '@testing-library/react';
import { useState } from 'react';
import ErrorBoundary from './ErrorBoundary';

// No i18n mock and no provider wrapper: useI18n() is resilient outside an
// I18nProvider (it echoes the key), so the real hook drives the fallback UI
// here and the assertions match the raw translation keys.
const ThrowingChild = ({ shouldThrow }: { shouldThrow: boolean }) => {
    if (shouldThrow) {
        throw new Error('boom');
    }

    return <div>Recovered child</div>;
};

const RecoverableHarness = () => {
    const [shouldThrow, setShouldThrow] = useState(true);

    return (
        <>
            <button type="button" onClick={() => setShouldThrow(false)}>
                Recover
            </button>
            <ErrorBoundary>
                <ThrowingChild shouldThrow={shouldThrow} />
            </ErrorBoundary>
        </>
    );
};

describe('components/ErrorBoundary', () => {
    let consoleErrorSpy: jest.SpyInstance;

    beforeEach(() => {
        consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => undefined);
    });

    afterEach(() => {
        consoleErrorSpy.mockRestore();
    });

    it('renders children when no error is thrown', () => {
        render(
            <ErrorBoundary>
                <div>Safe child</div>
            </ErrorBoundary>
        );

        expect(screen.getByText('Safe child')).toBeInTheDocument();
    });

    it('shows the fallback UI and can recover after reset', () => {
        render(<RecoverableHarness />);

        expect(screen.getByText('SOMETHING_WENT_WRONG')).toBeInTheDocument();
        expect(screen.getByText('boom')).toBeInTheDocument();
        expect(consoleErrorSpy).toHaveBeenCalledWith(
            'ErrorBoundary caught:',
            expect.objectContaining({ message: 'boom' }),
            expect.any(Object)
        );

        fireEvent.click(screen.getByRole('button', { name: 'Recover' }));
        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        expect(screen.getByText('Recovered child')).toBeInTheDocument();
    });
});
