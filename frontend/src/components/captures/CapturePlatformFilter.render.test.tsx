import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import CapturePlatformFilter from './CapturePlatformFilter';

describe('CapturePlatformFilter without mocks', () => {
    it('mounts against an absent backend and hides the select', () => {
        renderWithoutBackend(<CapturePlatformFilter selectedPlatform="" onChange={jest.fn()} />);

        expect(screen.queryByRole('combobox')).not.toBeInTheDocument();
    });

    it('keeps the platform from the URL selectable when the backend is absent', () => {
        renderWithoutBackend(
            <CapturePlatformFilter selectedPlatform="youtube" onChange={jest.fn()} />
        );

        expect(screen.getByRole('combobox')).toHaveTextContent('youtube');
    });
});
