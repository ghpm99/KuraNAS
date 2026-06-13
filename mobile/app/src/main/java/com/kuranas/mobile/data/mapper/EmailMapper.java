package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.EmailItem;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * Maps the {@code GET /api/v1/email/messages} paginated response into the lean
 * {@link EmailItem}s the kiosk renders. Only the {@code items} array is read; no
 * body field exists in the DTO by design (small payload for the old tablet).
 */
public final class EmailMapper {

    private EmailMapper() {
    }

    public static List<EmailItem> fromPaginatedJson(JSONObject json) throws JSONException {
        List<EmailItem> items = new ArrayList<EmailItem>();
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

    public static EmailItem fromJson(JSONObject json) {
        return new EmailItem(
                json.optString("sender_name", ""),
                json.optString("sender_address", ""),
                json.optString("subject", ""),
                json.optString("snippet", ""),
                json.optString("summary", ""),
                json.optString("importance", ""),
                json.optString("verdict", ""),
                json.optString("received_at", "")
        );
    }
}
