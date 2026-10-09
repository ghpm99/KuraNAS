import { render, screen } from '@testing-library/react';
import { ThemeProvider, createTheme } from '@mui/material';
import ColdTierIndicator from './coldTierIndicator';

const minimumGraphicContrastRatio = 3;
const lightPaperColor = '#ffffff';
const darkPaperColor = '#121212';

const toLinearChannel = (channel: number): number => {
    const normalized = channel / 255;
    return normalized <= 0.03928 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4;
};

const relativeLuminance = (hexColor: string): number => {
    const [red, green, blue] = [1, 3, 5].map((offset) =>
        toLinearChannel(parseInt(hexColor.slice(offset, offset + 2), 16))
    );
    return 0.2126 * red! + 0.7152 * green! + 0.0722 * blue!;
};

const contrastRatio = (foreground: string, background: string): number => {
    const [lighter, darker] = [relativeLuminance(foreground), relativeLuminance(background)].sort(
        (first, second) => second - first
    );
    return (lighter! + 0.05) / (darker! + 0.05);
};

const rgbToHex = (rgbColor: string): string => {
    const channels = rgbColor.match(/\d+/g)!.map(Number);
    return `#${channels.map((channel) => channel.toString(16).padStart(2, '0')).join('')}`;
};

const renderedIndicatorColor = (mode: 'light' | 'dark'): string => {
    render(
        <ThemeProvider theme={createTheme({ palette: { mode } })}>
            <ColdTierIndicator />
        </ThemeProvider>
    );
    return rgbToHex(getComputedStyle(screen.getByRole('img')).color);
};

describe('ColdTierIndicator contrast', () => {
    it('stays readable against the light paper background', () => {
        expect(
            contrastRatio(renderedIndicatorColor('light'), lightPaperColor)
        ).toBeGreaterThanOrEqual(minimumGraphicContrastRatio);
    });

    it('stays readable against the dark paper background', () => {
        expect(
            contrastRatio(renderedIndicatorColor('dark'), darkPaperColor)
        ).toBeGreaterThanOrEqual(minimumGraphicContrastRatio);
    });
});
