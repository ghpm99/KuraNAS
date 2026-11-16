import { readFileSync } from 'fs';
import path from 'path';
import { buildServerPage } from './buildServerPage';

interface GoldenItem {
    id: number;
    name: string;
}

const goldenDirectory = path.resolve(__dirname, '../../../../backend/pkg/utils/testdata');

function readGoldenEnvelope(fileName: string) {
    return JSON.parse(readFileSync(path.join(goldenDirectory, fileName), 'utf8'));
}

describe('buildServerPage', () => {
    it('matches the backend envelope of a page with a next page', () => {
        const goldenEnvelope = readGoldenEnvelope('pagination_page_with_next.json');

        expect(goldenEnvelope).toEqual(
            buildServerPage<GoldenItem>(goldenEnvelope.items, { page: 1, pageSize: 2, hasNext: true })
        );
        expect(goldenEnvelope.items).toHaveLength(2);
    });

    it('matches the backend envelope of the last page', () => {
        const goldenEnvelope = readGoldenEnvelope('pagination_last_page.json');

        expect(goldenEnvelope).toEqual(
            buildServerPage<GoldenItem>(goldenEnvelope.items, { page: 2, pageSize: 2, hasNext: false })
        );
        expect(goldenEnvelope.items).toHaveLength(1);
    });

    it.each(['pagination_page_with_next.json', 'pagination_last_page.json'])(
        'keeps %s in snake_case without camelCase pagination keys',
        (fileName) => {
            const goldenEnvelope = readGoldenEnvelope(fileName);

            expect(goldenEnvelope.pagination).not.toHaveProperty('hasNext');
            expect(goldenEnvelope.pagination).not.toHaveProperty('pageSize');
            expect(goldenEnvelope).not.toHaveProperty('hasNext');
            expect(goldenEnvelope).not.toHaveProperty('pageSize');
        }
    );
});
