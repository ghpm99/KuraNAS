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
import com.kuranas.mobile.domain.model.NotificationItem;

import java.util.ArrayList;
import java.util.List;

/** Renders the kiosk notification rows: a colour marker by type + title/message. */
public final class NotificationAdapter extends BaseAdapter {

    private final Context context;
    private final LayoutInflater inflater;
    private final List<NotificationItem> items = new ArrayList<NotificationItem>();

    public NotificationAdapter(Context context) {
        this.context = context;
        this.inflater = LayoutInflater.from(context);
    }

    public void setItems(List<NotificationItem> newItems) {
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
    public NotificationItem getItem(int position) {
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
            convertView = inflater.inflate(R.layout.item_kiosk_notification, parent, false);
            holder = new Holder();
            holder.marker = convertView.findViewById(R.id.marker);
            holder.title = (TextView) convertView.findViewById(R.id.notification_title);
            holder.message = (TextView) convertView.findViewById(R.id.notification_message);
            holder.time = (TextView) convertView.findViewById(R.id.notification_time);
            convertView.setTag(holder);
        } else {
            holder = (Holder) convertView.getTag();
        }

        NotificationItem item = items.get(position);
        holder.title.setText(item.getTitle());
        holder.message.setText(item.getMessage());
        holder.time.setText(TimeFormat.shortTime(item.getCreatedAt()));
        holder.marker.setBackgroundColor(ContextCompat.getColor(context, colorForType(item.getType())));

        return convertView;
    }

    private static int colorForType(String type) {
        if ("error".equals(type)) {
            return R.color.error;
        }
        if ("warning".equals(type)) {
            return R.color.warning;
        }
        if ("success".equals(type)) {
            return R.color.success;
        }
        return R.color.accent; // info / unknown
    }

    private static final class Holder {
        View marker;
        TextView title;
        TextView message;
        TextView time;
    }
}
