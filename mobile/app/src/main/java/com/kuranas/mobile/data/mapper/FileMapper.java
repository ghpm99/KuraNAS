package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.Pagination;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

public final class FileMapper {

    private FileMapper() {
    }

    public static FileItem fromJson(JSONObject json) throws JSONException {
        return new FileItem(
                json.optInt("id", 0),
                json.optString("name", ""),
                json.optString("path", ""),
                json.optString("parent_path", ""),
                json.optInt("type", 0),
                json.optString("format", ""),
                json.optLong("size", 0L),
                json.optString("updated_at", ""),
                json.optString("created_at", ""),
                json.optBoolean("starred", false),
                json.optInt("directory_content_count", 0)
        );
    }

    public static PaginatedResult<FileItem> fromPaginatedJson(JSONObject json) throws JSONException {
        JSONArray itemsArray = json.optJSONArray("items");
        List<FileItem> items = new ArrayList<FileItem>();
        if (itemsArray != null) {
            for (int i = 0; i < itemsArray.length(); i++) {
                items.add(fromJson(itemsArray.getJSONObject(i)));
            }
        }

        Pagination pagination = paginationFromJson(json.optJSONObject("pagination"));
        return new PaginatedResult<FileItem>(items, pagination);
    }

    public static Pagination paginationFromJson(JSONObject json) {
        if (json == null) {
            return new Pagination(1, 20, false, false);
        }
        return new Pagination(
                json.optInt("page", 1),
                json.optInt("page_size", 20),
                json.optBoolean("has_next", false),
                json.optBoolean("has_prev", false)
        );
    }
}
