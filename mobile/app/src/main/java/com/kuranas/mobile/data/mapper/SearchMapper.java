package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.SearchResult;
import com.kuranas.mobile.domain.model.VideoItem;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

public final class SearchMapper {

    private SearchMapper() {
    }

    public static SearchResult fromJson(JSONObject json) throws JSONException {
        String query = json.optString("query", "");
        String suggestion = json.optString("suggestion", "");

        List<FileItem> files = parseFileItems(json.optJSONArray("files"));
        List<FileItem> folders = parseFolderItems(json.optJSONArray("folders"));
        List<VideoItem> videos = parseVideoItems(json.optJSONArray("videos"));
        List<FileItem> images = parseFileItems(json.optJSONArray("images"));

        return new SearchResult(query, suggestion, files, folders, videos, images);
    }

    private static List<FileItem> parseFileItems(JSONArray array) throws JSONException {
        List<FileItem> items = new ArrayList<FileItem>();
        if (array == null) {
            return items;
        }
        for (int i = 0; i < array.length(); i++) {
            JSONObject obj = array.getJSONObject(i);
            items.add(new FileItem(
                    obj.optInt("id", 0),
                    obj.optString("name", ""),
                    obj.optString("path", ""),
                    obj.optString("parent_path", ""),
                    FileItem.TYPE_FILE,
                    obj.optString("format", ""),
                    0L,
                    "",
                    "",
                    obj.optBoolean("starred", false),
                    0
            ));
        }
        return items;
    }

    private static List<FileItem> parseFolderItems(JSONArray array) throws JSONException {
        List<FileItem> items = new ArrayList<FileItem>();
        if (array == null) {
            return items;
        }
        for (int i = 0; i < array.length(); i++) {
            JSONObject obj = array.getJSONObject(i);
            items.add(new FileItem(
                    obj.optInt("id", 0),
                    obj.optString("name", ""),
                    obj.optString("path", ""),
                    obj.optString("parent_path", ""),
                    FileItem.TYPE_DIRECTORY,
                    "",
                    0L,
                    "",
                    "",
                    obj.optBoolean("starred", false),
                    0
            ));
        }
        return items;
    }

    private static List<VideoItem> parseVideoItems(JSONArray array) throws JSONException {
        List<VideoItem> items = new ArrayList<VideoItem>();
        if (array == null) {
            return items;
        }
        for (int i = 0; i < array.length(); i++) {
            JSONObject obj = array.getJSONObject(i);
            items.add(new VideoItem(
                    obj.optInt("id", 0),
                    obj.optString("name", ""),
                    obj.optString("path", ""),
                    obj.optString("parent_path", ""),
                    obj.optString("format", ""),
                    0L,
                    "",
                    ""
            ));
        }
        return items;
    }
}
