import { render } from '@testing-library/react';
import DocumentFormatIcon from './DocumentFormatIcon';

describe('DocumentFormatIcon', () => {
    it.each(['.pdf', '.xlsx', '.pptx', '.txt', '.zzz', '', undefined])(
        'renders an icon for %s',
        (format) => {
            const { container } = render(<DocumentFormatIcon format={format as string} />);

            expect(container.querySelector('svg')).not.toBeNull();
        }
    );
});
