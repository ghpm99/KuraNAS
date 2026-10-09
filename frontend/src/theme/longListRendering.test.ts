import { readFileSync } from 'fs';
import { join } from 'path';

const readSource = (relativePath: string) =>
    readFileSync(join(__dirname, '..', relativePath), 'utf8');

describe('long list item rendering cost', () => {
    it.each([
        ['features/files/fileListRow/fileListRow.module.css', '--app-intrinsic-file-row-height'],
        ['components/imageContent/ImageContent.module.css', '--app-intrinsic-image-tile-height'],
        [
            'features/videos/components/videoContent/videoContent.module.css',
            '--app-intrinsic-video-card-height',
        ],
    ])('defers offscreen rendering in %s', (stylesheetPath, intrinsicSizeVariable) => {
        const stylesheetSource = readSource(stylesheetPath);

        expect(stylesheetSource).toContain('content-visibility: auto;');
        expect(stylesheetSource).toContain(
            `contain-intrinsic-block-size: auto var(${intrinsicSizeVariable});`
        );
    });

    it.each([
        ['features/files/fileCard/fileCard.tsx', '--app-intrinsic-file-card-height'],
        ['features/music/components/TrackListItem.tsx', '--app-intrinsic-track-row-height'],
    ])('defers offscreen rendering in %s', (componentPath, intrinsicSizeVariable) => {
        const componentSource = readSource(componentPath);

        expect(componentSource).toContain("contentVisibility: 'auto'");
        expect(componentSource).toContain(`auto var(${intrinsicSizeVariable})`);
    });
});
