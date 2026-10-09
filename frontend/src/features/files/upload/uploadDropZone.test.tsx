import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import FilesLayout from '../files/filesLayout';
import UploadDropZone from './uploadDropZone';

describe('UploadDropZone without mocks', () => {
    it('renders its children inside the real files providers while the backend is absent', () => {
        render(
            <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
                <MemoryRouter>
                    <FilesLayout>
                        <UploadDropZone>
                            <span>listing</span>
                        </UploadDropZone>
                    </FilesLayout>
                </MemoryRouter>
            </QueryClientProvider>
        );

        expect(screen.getByText('listing')).toBeInTheDocument();
        expect(screen.queryByText('FILES_UPLOAD_DROP_HINT')).not.toBeInTheDocument();
    });
});
