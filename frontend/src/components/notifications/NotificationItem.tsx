import { Box, Typography } from '@mui/material';
import { AlertCircle, CheckCircle, Info, AlertTriangle, Monitor } from 'lucide-react';
import type { Notification, NotificationType } from '@/types/notification';

const typeConfig: Record<NotificationType, { icon: typeof Info; color: string }> = {
    info: { icon: Info, color: '#2196f3' },
    success: { icon: CheckCircle, color: '#4caf50' },
    warning: { icon: AlertTriangle, color: '#ff9800' },
    error: { icon: AlertCircle, color: '#f44336' },
    system: { icon: Monitor, color: '#9e9e9e' },
};

function formatRelativeTime(dateStr: string): string {
    const now = new Date();
    const date = new Date(dateStr);
    const diffMs = now.getTime() - date.getTime();
    const diffMinutes = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMinutes / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMinutes < 1) return 'now';
    if (diffMinutes < 60) return `${diffMinutes}m`;
    if (diffHours < 24) return `${diffHours}h`;
    if (diffDays < 30) return `${diffDays}d`;
    return date.toLocaleDateString();
}

interface NotificationItemProps {
    notification: Notification;
    onClick?: () => void;
}

export default function NotificationItem({ notification, onClick }: NotificationItemProps) {
    const config = typeConfig[notification.type] ?? typeConfig.info;
    const Icon = config.icon;

    return (
        <Box
            onClick={onClick}
            sx={{
                display: 'flex',
                gap: 1.5,
                p: 1.5,
                cursor: onClick ? 'pointer' : 'default',
                opacity: notification.is_read ? 0.6 : 1,
                borderRadius: 1,
                '&:hover': onClick
                    ? { bgcolor: 'rgba(255,255,255,0.04)' }
                    : undefined,
            }}
        >
            <Box sx={{ mt: 0.25, flexShrink: 0 }}>
                <Icon size={18} color={config.color} />
            </Box>
            <Box sx={{ flex: 1, minWidth: 0 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Typography
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
                            variant="caption"
                            sx={{
                                bgcolor: 'rgba(255,255,255,0.08)',
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
                <Typography variant="caption" sx={{ color: 'text.disabled', fontSize: '0.65rem' }}>
                    {formatRelativeTime(notification.created_at)}
                </Typography>
            </Box>
            {!notification.is_read && (
                <Box
                    sx={{
                        width: 8,
                        height: 8,
                        borderRadius: '50%',
                        bgcolor: 'var(--app-color-primary)',
                        flexShrink: 0,
                        mt: 0.75,
                    }}
                />
            )}
        </Box>
    );
}
