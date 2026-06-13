package com.kuranas.mobile.presentation.home;

import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ListView;
import android.widget.TextView;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.data.remote.api.EmailApi;
import com.kuranas.mobile.data.remote.api.NotificationApi;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.EmailItem;
import com.kuranas.mobile.domain.model.NotificationItem;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.presentation.base.BaseFragment;

import java.text.SimpleDateFormat;
import java.util.Calendar;
import java.util.List;
import java.util.Locale;

/**
 * The single screen of the wall-panel app: a giant clock + date on top, and two
 * read-only panels below (notifications | e-mails) kept fresh by staggered
 * polling with exponential backoff and an offline indicator. No interaction.
 */
public class HomeFragment extends BaseFragment {

    private static final long CLOCK_INTERVAL_MS = 1000;

    private static final int PAGE_SIZE = 8;
    private static final long NOTIF_BASE_MS = 60_000;
    private static final long EMAIL_BASE_MS = 120_000;
    private static final long BACKOFF_MAX_MS = 300_000;
    // Offset the e-mail poll from the notifications poll so requests don't coincide.
    private static final long EMAIL_START_OFFSET_MS = 30_000;

    private TextView clockText;
    private TextView dateText;

    private TextView notificationsOffline;
    private TextView emailsOffline;
    private NotificationAdapter notificationAdapter;
    private EmailAdapter emailAdapter;

    private NotificationApi notificationApi;
    private EmailApi emailApi;

    private final Backoff notificationBackoff = new Backoff(NOTIF_BASE_MS, BACKOFF_MAX_MS);
    private final Backoff emailBackoff = new Backoff(EMAIL_BASE_MS, BACKOFF_MAX_MS);

    private final Handler handler = new Handler(Looper.getMainLooper());
    private boolean active;

    private final Runnable clockRunnable = new Runnable() {
        @Override
        public void run() {
            updateClock();
            handler.postDelayed(this, CLOCK_INTERVAL_MS);
        }
    };

    private final Runnable notificationPoll = new Runnable() {
        @Override
        public void run() {
            pollNotifications();
        }
    };

    private final Runnable emailPoll = new Runnable() {
        @Override
        public void run() {
            pollEmails();
        }
    };

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_home, container, false);

        clockText = (TextView) root.findViewById(R.id.clock_text);
        dateText = (TextView) root.findViewById(R.id.date_text);

        notificationsOffline = (TextView) root.findViewById(R.id.notifications_offline);
        emailsOffline = (TextView) root.findViewById(R.id.emails_offline);
        notificationsOffline.setText(t("KIOSK_OFFLINE"));
        emailsOffline.setText(t("KIOSK_OFFLINE"));

        ((TextView) root.findViewById(R.id.notifications_title)).setText(t("KIOSK_NOTIFICATIONS_TITLE"));
        ((TextView) root.findViewById(R.id.emails_title)).setText(t("KIOSK_EMAILS_TITLE"));

        TextView notificationsEmpty = (TextView) root.findViewById(R.id.notifications_empty);
        TextView emailsEmpty = (TextView) root.findViewById(R.id.emails_empty);
        notificationsEmpty.setText(t("KIOSK_EMPTY_NOTIFICATIONS"));
        emailsEmpty.setText(t("KIOSK_EMPTY_EMAILS"));

        notificationAdapter = new NotificationAdapter(requireContext());
        emailAdapter = new EmailAdapter(requireContext(), getTranslations());

        ListView notificationsList = (ListView) root.findViewById(R.id.notifications_list);
        ListView emailsList = (ListView) root.findViewById(R.id.emails_list);
        notificationsList.setAdapter(notificationAdapter);
        emailsList.setAdapter(emailAdapter);
        notificationsList.setEmptyView(notificationsEmpty);
        emailsList.setEmptyView(emailsEmpty);

        notificationApi = ServiceLocator.getInstance().getNotificationApi();
        emailApi = ServiceLocator.getInstance().getEmailApi();

        updateClock();
        return root;
    }

    @Override
    public void onResume() {
        super.onResume();
        active = true;

        updateClock();
        handler.postDelayed(clockRunnable, CLOCK_INTERVAL_MS);

        // Fetch both promptly, but stagger so the two requests don't coincide.
        handler.post(notificationPoll);
        handler.postDelayed(emailPoll, EMAIL_START_OFFSET_MS);
    }

    @Override
    public void onPause() {
        super.onPause();
        active = false;
        handler.removeCallbacksAndMessages(null);
    }

    private void updateClock() {
        Calendar now = Calendar.getInstance();
        SimpleDateFormat timeFormat = new SimpleDateFormat("HH:mm:ss", Locale.getDefault());
        SimpleDateFormat dateFormat = new SimpleDateFormat("EEEE, d 'de' MMMM", new Locale("pt", "BR"));
        clockText.setText(timeFormat.format(now.getTime()));
        dateText.setText(dateFormat.format(now.getTime()));
    }

    private void pollNotifications() {
        if (!active || notificationApi == null) {
            return;
        }
        notificationApi.getRecent(PAGE_SIZE, new ApiCallback<List<NotificationItem>>() {
            @Override
            public void onSuccess(List<NotificationItem> result) {
                if (!active) {
                    return;
                }
                notificationAdapter.setItems(result);
                notificationsOffline.setVisibility(View.GONE);
                notificationBackoff.recordSuccess();
                scheduleNotifications(notificationBackoff.currentDelayMs());
            }

            @Override
            public void onError(AppError error) {
                if (!active) {
                    return;
                }
                notificationsOffline.setVisibility(View.VISIBLE);
                scheduleNotifications(notificationBackoff.recordFailure());
            }
        });
    }

    private void pollEmails() {
        if (!active || emailApi == null) {
            return;
        }
        emailApi.getRecent(PAGE_SIZE, new ApiCallback<List<EmailItem>>() {
            @Override
            public void onSuccess(List<EmailItem> result) {
                if (!active) {
                    return;
                }
                emailAdapter.setItems(result);
                emailsOffline.setVisibility(View.GONE);
                emailBackoff.recordSuccess();
                scheduleEmails(emailBackoff.currentDelayMs());
            }

            @Override
            public void onError(AppError error) {
                if (!active) {
                    return;
                }
                emailsOffline.setVisibility(View.VISIBLE);
                scheduleEmails(emailBackoff.recordFailure());
            }
        });
    }

    private void scheduleNotifications(long delayMs) {
        handler.removeCallbacks(notificationPoll);
        handler.postDelayed(notificationPoll, delayMs);
    }

    private void scheduleEmails(long delayMs) {
        handler.removeCallbacks(emailPoll);
        handler.postDelayed(emailPoll, delayMs);
    }
}
