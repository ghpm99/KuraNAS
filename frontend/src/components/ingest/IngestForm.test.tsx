import { fireEvent, screen, waitFor } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import IngestForm from './IngestForm';

describe('IngestForm without backend', () => {
    it('renders without a backend', async () => {
        expectRendersWithoutBackend(<IngestForm />);
        await waitFor(() =>
            expect(screen.getByRole('button', { name: /INGEST_FORM_SUBMIT/ })).toBeDefined()
        );
    });

    it('keeps the submit button disabled while targets and presets are unavailable', () => {
        expectRendersWithoutBackend(<IngestForm />);
        const submitButton = screen.getByRole('button', { name: /INGEST_FORM_SUBMIT/ });
        expect((submitButton as HTMLButtonElement).disabled).toBe(true);
        fireEvent.click(submitButton);
    });
});
