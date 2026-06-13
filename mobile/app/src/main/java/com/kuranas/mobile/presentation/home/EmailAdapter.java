package com.kuranas.mobile.presentation.home;

import android.content.Context;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.BaseAdapter;
import android.widget.TextView;

import androidx.core.content.ContextCompat;

import com.kuranas.mobile.R;
import com.kuranas.mobile.data.mapper.TimeFormat;
import com.kuranas.mobile.domain.model.EmailItem;
import com.kuranas.mobile.i18n.TranslationManager;

import java.util.ArrayList;
import java.util.List;

/**
 * Renders the kiosk e-mail rows. Flagged messages (malicious/suspicious) get a
 * red/amber marker and never show the AI summary; high-importance safe messages
 * get a blue marker and an "important" badge. No HTML is ever rendered.
 */
public final class EmailAdapter extends BaseAdapter {

    private final Context context;
    private final LayoutInflater inflater;
    private final TranslationManager translations;
    private final List<EmailItem> items = new ArrayList<EmailItem>();

    public EmailAdapter(Context context, TranslationManager translations) {
        this.context = context;
        this.inflater = LayoutInflater.from(context);
        this.translations = translations;
    }

    public void setItems(List<EmailItem> newItems) {
        items.clear();
        if (newItems != null) {
            items.addAll(newItems);
        }
        notifyDataSetChanged();
    }

    @Override
    public int getCount() {
        return items.size();
    }

    @Override
    public EmailItem getItem(int position) {
        return items.get(position);
    }

    @Override
    public long getItemId(int position) {
        return position;
    }

    @Override
    public View getView(int position, View convertView, ViewGroup parent) {
        Holder holder;
        if (convertView == null) {
            convertView = inflater.inflate(R.layout.item_kiosk_email, parent, false);
            holder = new Holder();
            holder.marker = convertView.findViewById(R.id.marker);
            holder.sender = (TextView) convertView.findViewById(R.id.email_sender);
            holder.badge = (TextView) convertView.findViewById(R.id.email_badge);
            holder.time = (TextView) convertView.findViewById(R.id.email_time);
            holder.subject = (TextView) convertView.findViewById(R.id.email_subject);
            holder.summary = (TextView) convertView.findViewById(R.id.email_summary);
            convertView.setTag(holder);
        } else {
            holder = (Holder) convertView.getTag();
        }

        EmailItem item = items.get(position);
        holder.sender.setText(senderLabel(item));
        holder.time.setText(TimeFormat.shortTime(item.getReceivedAt()));
        holder.subject.setText(item.getSubject());
        holder.marker.setBackgroundColor(ContextCompat.getColor(context, markerColor(item)));

        // Summary only for non-flagged messages that actually have one.
        if (!item.isFlagged() && item.getSummary() != null && !item.getSummary().isEmpty()) {
            holder.summary.setText(item.getSummary());
            holder.summary.setVisibility(View.VISIBLE);
        } else {
            holder.summary.setVisibility(View.GONE);
        }

        // "Important" badge for high-importance, non-flagged messages.
        if (item.isHighImportance() && !item.isFlagged()) {
            holder.badge.setText(t("KIOSK_IMPORTANCE_HIGH"));
            holder.badge.setVisibility(View.VISIBLE);
        } else {
            holder.badge.setVisibility(View.GONE);
        }

        return convertView;
    }

    private String senderLabel(EmailItem item) {
        if (item.getSenderName() != null && !item.getSenderName().isEmpty()) {
            return item.getSenderName();
        }
        return item.getSenderAddress();
    }

    private static int markerColor(EmailItem item) {
        if ("malicious".equals(item.getVerdict())) {
            return R.color.error;
        }
        if ("suspicious".equals(item.getVerdict())) {
            return R.color.warning;
        }
        if (item.isHighImportance()) {
            return R.color.accent;
        }
        return R.color.surface;
    }

    private String t(String key) {
        return translations != null ? translations.t(key) : key;
    }

    private static final class Holder {
        View marker;
        TextView sender;
        TextView badge;
        TextView time;
        TextView subject;
        TextView summary;
    }
}
