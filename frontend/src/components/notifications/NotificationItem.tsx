import { Box, ButtonBase, Typography } from '@mui/material';
import { AlertCircle, CheckCircle, Info, AlertTriangle, Monitor } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { Notification, NotificationType } from '@/types/notification';
import { formatNotificationRelativeTime } from './notificationRelativeTime';

const typeConfig: Record<NotificationType, { icon: typeof Info; colorToken: string }> = {
    info: { icon: Info, colorToken: 'var(--app-color-link)' },
    success: { icon: CheckCircle, colorToken: 'var(--app-color-success)' },
    warning: { icon: AlertTriangle, colorToken: 'var(--app-color-warning)' },
    error: { icon: AlertCircle, colorToken: 'var(--app-color-danger)' },
    system: { icon: Monitor, colorToken: 'var(--app-color-text-muted)' },
};

const visuallyHiddenSx = {
    border: 0,
    clip: 'rect(0 0 0 0)',
    height: '1px',
    margin: '-1px',
    overflow: 'hidden',
    padding: 0,
    position: 'absolute',
    whiteSpace: 'nowrap',
    width: '1px',
} as const;

interface NotificationItemProps {
    notification: Notification;
    onClick?: () => void;
}

export default function NotificationItem({ notification, onClick }: NotificationItemProps) {
    const { t } = useI18n();
    const config = typeConfig[notification.type] ?? typeConfig.info;
    const Icon = config.icon;

    const content = (
        <>
            <Box
                component="span"
                sx={{ mt: 0.25, flexShrink: 0, display: 'flex', color: config.colorToken }}
            >
                <Icon size={18} aria-hidden="true" />
            </Box>
            <Box component="span" sx={{ flex: 1, minWidth: 0, display: 'block' }}>
                <Box component="span" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Typography
                        component="span"
                        variant="body2"
                        sx={{
                            fontWeight: notification.is_read ? 400 : 600,
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                            flex: 1,
                        }}
                    >
                        {notification.title}
                    </Typography>
                    {notification.is_grouped && notification.group_count > 1 && (
                        <Typography
                            component="span"
                            variant="caption"
                            sx={{
                                bgcolor: 'rgba(var(--app-color-ink-rgb), 0.08)',
                                px: 0.75,
                                py: 0.125,
                                borderRadius: 1,
                                fontSize: '0.7rem',
                                flexShrink: 0,
                            }}
                        >
                            x{notification.group_count}
                        </Typography>
                    )}
                </Box>
                <Typography
                    component="span"
                    variant="caption"
                    sx={{
                        color: 'text.secondary',
                        display: 'block',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                    }}
                >
                    {notification.message}
                </Typography>
                <Typography
                    component="span"
                    variant="caption"
                    sx={{ color: 'text.disabled', fontSize: '0.65rem', display: 'block' }}
                >
                    {formatNotificationRelativeTime(notification.created_at)}
                </Typography>
            </Box>
            {!notification.is_read && (
                <Box
                    component="span"
                    sx={{
                        width: 8,
                        height: 8,
                        borderRadius: '50%',
                        bgcolor: 'var(--app-color-primary)',
                        flexShrink: 0,
                        mt: 0.75,
                    }}
                >
                    <Box component="span" sx={visuallyHiddenSx}>
                        {t('UNREAD')}
                    </Box>
                </Box>
            )}
        </>
    );

    const sharedSx = {
        display: 'flex',
        gap: 1.5,
        p: 1.5,
        width: '100%',
        textAlign: 'left',
        opacity: notification.is_read ? 0.6 : 1,
        borderRadius: 1,
    } as const;

    if (!onClick) {
        return <Box sx={sharedSx}>{content}</Box>;
    }

    return (
        <ButtonBase
            onClick={onClick}
            sx={{
                ...sharedSx,
                justifyContent: 'flex-start',
                alignItems: 'flex-start',
                '&:hover': { bgcolor: 'rgba(var(--app-color-ink-rgb), 0.04)' },
                '&.Mui-focusVisible': {
                    outline: '2px solid var(--app-color-primary)',
                    outlineOffset: -2,
                },
            }}
        >
            {content}
        </ButtonBase>
    );
}
