import { useMediaQuery, useTheme } from '@mui/material';

export type ActionBarLayout = {
    isIconOnly: boolean;
    isSecondaryCollapsed: boolean;
};

const useActionBarLayout = (): ActionBarLayout => {
    const theme = useTheme();
    const isIconOnly = useMediaQuery(theme.breakpoints.down('md'));
    const isSecondaryCollapsed = useMediaQuery(theme.breakpoints.down('sm'));

    return { isIconOnly, isSecondaryCollapsed };
};

export default useActionBarLayout;
