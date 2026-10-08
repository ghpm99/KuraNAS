export const supportsProgrammaticVolume = (): boolean => {
    if (typeof document === 'undefined') return false;
    const probeElement = document.createElement('audio');
    probeElement.volume = 0.5;
    return probeElement.volume === 0.5;
};
