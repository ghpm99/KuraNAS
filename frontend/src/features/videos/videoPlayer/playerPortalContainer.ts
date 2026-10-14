import { createContext, useContext } from 'react';

export const PlayerPortalContainerContext = createContext<HTMLElement | null>(null);

export const usePlayerPortalContainer = () => useContext(PlayerPortalContainerContext);
