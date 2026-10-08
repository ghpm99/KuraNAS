import { Box, Button, CircularProgress, Tab, Tabs } from '@mui/material';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Bell } from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import EmptyState from '@/components/emptyState/emptyState';
import ErrorState from '@/components/errorState/errorState';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useI18n from '@/components/i18n/provider/i18nContext';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import {
    getNotifications,
    markAllNotificationsAsRead,
    markNotificationAsRead,
} from '@/service/notifications';
import type { Notification, NotificationType } from '@/types/notification';
import NotificationItem from './NotificationItem';
import { resolveNotificationTargetRoute } from './notificationTargetRoute';

type FilterTab = 'all' | 'unread' | NotificationType;

const PAGE_SIZE = 20;

const buildFilterParams = (activeTab: FilterTab) => {
    if (activeTab === 'all') return {};
    if (activeTab === 'unread') return { is_read: false };
    return { type: activeTab };
};

export default function NotificationsScreen() {
    const { t } = useI18n();
    const navigate = useNavigate();
    const queryClient = useQueryClient();
    const [activeTab, setActiveTab] = useState<FilterTab>('all');

    const {
        data,
        isLoading,
        isError,
        error,
        refetch,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useInfiniteQuery({
        queryKey: ['notifications-page', activeTab],
        queryFn: ({ pageParam }) =>
            getNotifications({
                page: pageParam,
                pageSize: PAGE_SIZE,
                ...buildFilterParams(activeTab),
            }),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage?.pagination?.has_next ? lastPage.pagination.page + 1 : undefined,
    });

    const invalidateNotificationQueries = async () => {
        await Promise.all([
            queryClient.invalidateQueries({ queryKey: ['notifications-page'] }),
            queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] }),
            queryClient.invalidateQueries({ queryKey: ['notifications'] }),
        ]);
    };

    const markOneMutation = useMutation({
        mutationFn: markNotificationAsRead,
        onSuccess: invalidateNotificationQueries,
    });

    const markAllMutation = useMutation({
        mutationFn: markAllNotificationsAsRead,
        onSuccess: invalidateNotificationQueries,
    });

    const allNotifications = data?.pages.flatMap((page) => page?.items ?? []) ?? [];

    const handleItemClick = (notification: Notification) => {
        const targetRoute = resolveNotificationTargetRoute(notification);
        if (!notification.is_read) {
            markOneMutation.mutate(notification.id);
        }
        if (targetRoute) {
            navigate(targetRoute);
        }
    };

    const renderBody = () => {
        if (isLoading) {
            return (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                    <CircularProgress size={24} role="progressbar" aria-label={t('LOADING')} />
                </Box>
            );
        }
        if (isError) {
            return (
                <ErrorState
                    title={t('NOTIFICATIONS_LOAD_ERROR')}
                    backendMessage={extractBackendErrorMessage(error)}
                    onRetry={() => void refetch()}
                />
            );
        }
        if (allNotifications.length === 0) {
            return (
                <EmptyState
                    title={t('NO_NOTIFICATIONS')}
                    icon={<Bell size={32} aria-hidden="true" />}
                />
            );
        }
        return (
            <>
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                    {allNotifications.map((notification) => (
                        <NotificationItem
                            key={notification.id}
                            notification={notification}
                            onClick={() => handleItemClick(notification)}
                        />
                    ))}
                </Box>
                <LoadMoreSentinel
                    hasNextPage={hasNextPage}
                    isFetchingNextPage={isFetchingNextPage}
                    fetchNextPage={() => void fetchNextPage()}
                />
            </>
        );
    };

    return (
        <PageContainer>
            <PageHeader
                title={t('NOTIFICATIONS')}
                subtitle={t('NOTIFICATIONS_SUBTITLE')}
                actions={
                    <Button
                        size="small"
                        onClick={() => markAllMutation.mutate()}
                        disabled={markAllMutation.isPending}
                    >
                        {t('MARK_ALL_AS_READ')}
                    </Button>
                }
            />

            <Tabs
                value={activeTab}
                onChange={(_, selectedTab: FilterTab) => setActiveTab(selectedTab)}
                variant="scrollable"
                scrollButtons="auto"
                aria-label={t('NOTIFICATIONS')}
                sx={{ minHeight: 36, '& .MuiTab-root': { minHeight: 36, py: 0.5 } }}
            >
                <Tab label={t('ALL')} value="all" />
                <Tab label={t('UNREAD')} value="unread" />
                <Tab label={t('NOTIFICATION_TYPE_INFO')} value="info" />
                <Tab label={t('NOTIFICATION_TYPE_SUCCESS')} value="success" />
                <Tab label={t('NOTIFICATION_TYPE_WARNING')} value="warning" />
                <Tab label={t('NOTIFICATION_TYPE_ERROR')} value="error" />
            </Tabs>

            {renderBody()}
        </PageContainer>
    );
}
