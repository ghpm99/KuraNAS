import type { MouseEvent } from 'react';

export const isNewTabClick = (event: MouseEvent<HTMLElement>): boolean =>
    event.button !== 0 || event.ctrlKey || event.metaKey;
