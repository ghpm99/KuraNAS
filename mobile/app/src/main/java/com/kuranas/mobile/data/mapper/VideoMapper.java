package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.Pagination;
import com.kuranas.mobile.domain.model.VideoItem;
import com.kuranas.mobile.domain.model.VideoPlaybackState;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

public final class VideoMapper {

    private VideoMapper() {
    }

    public static VideoItem fromJson(JSONObject json) throws JSONException {
        return new VideoItem(
                json.optInt("id", 0),
                json.optString("name", ""),
                json.optString("path", ""),
                json.optString("parent_path", ""),
                json.optString("format", ""),
                json.optLong("size", 0L),
                json.optString("created_at", ""),
                json.optString("updated_at", "")
        );
    }

    public static PaginatedResult<VideoItem> fromPaginatedJson(JSONObject json) throws JSONException {
        JSONArray itemsArray = json.optJSONArray("items");
        List<VideoItem> items = new ArrayList<VideoItem>();
        if (itemsArray != null) {
            for (int i = 0; i < itemsArray.length(); i++) {
                items.add(fromJson(itemsArray.getJSONObject(i)));
            }
        }

        Pagination pagination = FileMapper.paginationFromJson(json.optJSONObject("pagination"));
        return new PaginatedResult<VideoItem>(items, pagination);
    }

    public static VideoPlaybackState playbackStateFromJson(JSONObject json) throws JSONException {
        return new VideoPlaybackState(
                json.optInt("id", 0),
                json.optString("client_id", ""),
                json.optInt("playlist_id", 0),
                json.optInt("video_id", 0),
                json.optDouble("current_time", 0.0),
                json.optDouble("duration", 0.0),
                json.optBoolean("is_paused", false),
                json.optBoolean("completed", false),
                json.optString("last_update", "")
        );
    }
}
