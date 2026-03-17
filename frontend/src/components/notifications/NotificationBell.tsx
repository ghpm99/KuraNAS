import { Badge, IconButton, Popover } from '@mui/material';
import { Bell } from 'lucide-react';
import { useCallback, useState } from 'react';
import { useNotifications } from '@/components/providers/notificationProvider/notificationContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import NotificationDropdown from './NotificationDropdown';

interface NotificationBellProps {
    className?: string;
}

export default function NotificationBell({ className }: NotificationBellProps) {
    const { t } = useI18n();
    const { unreadCount } = useNotifications();
    const [anchorEl, setAnchorEl] = useState<HTMLButtonElement | null>(null);

    const handleClick = useCallback((event: React.MouseEvent<HTMLButtonElement>) => {
        setAnchorEl((prev) => (prev ? null : event.currentTarget));
    }, []);

    const handleClose = useCallback(() => {
        setAnchorEl(null);
    }, []);

    return (
        <>
            <IconButton
                title={t('NOTIFICATIONS')}
                aria-label={t('NOTIFICATIONS')}
                size="small"
                className={className}
                onClick={handleClick}
            >
                <Badge
                    badgeContent={unreadCount}
                    color="error"
                    max={99}
                    invisible={unreadCount === 0}
                    sx={{
                        '& .MuiBadge-badge': {
                            fontSize: '0.65rem',
                            minWidth: 16,
                            height: 16,
                        },
                    }}
                >
                    <Bell size={16} />
                </Badge>
            </IconButton>
            <Popover
                open={Boolean(anchorEl)}
                anchorEl={anchorEl}
                onClose={handleClose}
                anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
                transformOrigin={{ vertical: 'top', horizontal: 'right' }}
                slotProps={{
                    paper: {
                        sx: {
                            mt: 1,
                            bgcolor: '#1e1e2e',
                            backgroundImage: 'none',
                            border: '1px solid rgba(255,255,255,0.08)',
                            borderRadius: 2,
                        },
                    },
                }}
            >
                <NotificationDropdown onClose={handleClose} />
            </Popover>
        </>
    );
}
