import { useCallback, useContext } from 'react';
import { UNSAFE_NavigationContext } from 'react-router-dom';

export default function useOptionalNavigate() {
    const navigator = useContext(UNSAFE_NavigationContext)?.navigator;
    return useCallback((destination: string) => navigator?.push(destination), [navigator]);
}
