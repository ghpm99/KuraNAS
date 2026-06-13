package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.NotificationItem;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * Maps the {@code GET /api/v1/notifications} paginated response into the lean
 * {@link NotificationItem}s the kiosk renders. Only the {@code items} array is
 * read; the panel does not paginate.
 */
public final class NotificationMapper {

    private NotificationMapper() {
    }

    public static List<NotificationItem> fromPaginatedJson(JSONObject json) throws JSONException {
        List<NotificationItem> items = new ArrayList<NotificationItem>();
        if (json == null) {
            return items;
        }
        JSONArray array = json.optJSONArray("items");
        if (array == null) {
            return items;
        }
        for (int i = 0; i < array.length(); i++) {
            JSONObject obj = array.optJSONObject(i);
            if (obj != null) {
                items.add(fromJson(obj));
            }
        }
        return items;
    }

    public static NotificationItem fromJson(JSONObject json) {
        return new NotificationItem(
                json.optString("type", ""),
                json.optString("title", ""),
                json.optString("message", ""),
                json.optString("created_at", "")
        );
    }
}
