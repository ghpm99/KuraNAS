import { Box, Button, Divider, Typography } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { appRoutes } from '@/app/routes';
import { useNotifications } from '@/components/providers/notificationProvider/notificationContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import NotificationItem from './NotificationItem';

interface NotificationDropdownProps {
    onClose: () => void;
}

export default function NotificationDropdown({ onClose }: NotificationDropdownProps) {
    const { t } = useI18n();
    const navigate = useNavigate();
    const { notifications, unreadCount, markAsRead, markAllAsRead } = useNotifications();

    const handleViewAll = () => {
        onClose();
        navigate(appRoutes.notifications);
    };

    const handleItemClick = async (id: number, isRead: boolean) => {
        if (!isRead) {
            await markAsRead(id);
        }
    };

    return (
        <Box sx={{ width: 360, maxHeight: 440, display: 'flex', flexDirection: 'column' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', p: 2, pb: 1 }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 600 }}>
                    {t('NOTIFICATIONS')}
                </Typography>
                {unreadCount > 0 && (
                    <Button size="small" onClick={markAllAsRead} sx={{ textTransform: 'none', fontSize: '0.75rem' }}>
                        {t('MARK_ALL_AS_READ')}
                    </Button>
                )}
            </Box>
            <Divider />
            <Box sx={{ flex: 1, overflow: 'auto', py: 0.5 }}>
                {notifications.length === 0 ? (
                    <Box sx={{ p: 3, textAlign: 'center' }}>
                        <Typography variant="body2" color="text.secondary">
                            {t('NO_NOTIFICATIONS')}
                        </Typography>
                    </Box>
                ) : (
                    notifications.map((notification) => (
                        <NotificationItem
                            key={notification.id}
                            notification={notification}
                            onClick={() => handleItemClick(notification.id, notification.is_read)}
                        />
                    ))
                )}
            </Box>
            <Divider />
            <Box sx={{ p: 1, textAlign: 'center' }}>
                <Button size="small" onClick={handleViewAll} sx={{ textTransform: 'none', fontSize: '0.75rem' }}>
                    {t('VIEW_ALL_NOTIFICATIONS')}
                </Button>
            </Box>
        </Box>
    );
}
