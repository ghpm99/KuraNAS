export type RgbColor = readonly [number, number, number];

const goldenAngleDegrees = 137.508;
const idleSaturation = 0.5;
const idleLightness = 0.76;
const selectedSaturation = 0.85;
const selectedLightness = 0.5;

const hslChannel = (hue: number, saturation: number, lightness: number, offset: number): number => {
    const chroma = saturation * Math.min(lightness, 1 - lightness);
    const sector = (offset + hue / 30) % 12;
    const channel = lightness - chroma * Math.max(-1, Math.min(sector - 3, 9 - sector, 1));
    return Math.round(channel * 255);
};

/** Stable color for a room id; selected rooms are more saturated and darker. */
export const vacuumRoomColor = (roomId: number, isSelected: boolean): RgbColor => {
    const hue = (roomId * goldenAngleDegrees) % 360;
    const saturation = isSelected ? selectedSaturation : idleSaturation;
    const lightness = isSelected ? selectedLightness : idleLightness;
    return [
        hslChannel(hue, saturation, lightness, 0),
        hslChannel(hue, saturation, lightness, 8),
        hslChannel(hue, saturation, lightness, 4),
    ];
};

export const toCssColor = ([red, green, blue]: RgbColor): string => `rgb(${red}, ${green}, ${blue})`;
