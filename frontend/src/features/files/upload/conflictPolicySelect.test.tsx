import { fireEvent, render, screen, within } from '@testing-library/react';
import ConflictPolicySelect from './conflictPolicySelect';
import { UploadQueueContextProvider } from './uploadQueueContext';
import type { UploadQueue } from './uploadQueueTypes';

describe('ConflictPolicySelect', () => {
    it('renders without a provider showing the default policy', () => {
        render(<ConflictPolicySelect />);

        expect(screen.getByText('FILES_UPLOAD_ON_CONFLICT_RENAME')).toBeInTheDocument();
    });

    it('changes the policy through the queue context', () => {
        const setConflictPolicy = jest.fn();
        const queue = { conflictPolicy: 'rename', setConflictPolicy } as unknown as UploadQueue;
        render(
            <UploadQueueContextProvider value={queue}>
                <ConflictPolicySelect />
            </UploadQueueContextProvider>
        );

        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(within(screen.getByRole('listbox')).getByText('FILES_UPLOAD_ON_CONFLICT_SKIP'));

        expect(setConflictPolicy).toHaveBeenCalledWith('skip');
    });
});
