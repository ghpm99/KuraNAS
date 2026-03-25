package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.MusicPlayerState;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.Pagination;
import com.kuranas.mobile.domain.model.Track;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

public final class MusicMapper {

    private MusicMapper() {
    }

    public static Track trackFromJson(JSONObject json) throws JSONException {
        int id = json.optInt("id", 0);
        int position = json.optInt("position", 0);
        String addedAt = json.optString("added_at", "");

        JSONObject file = json.optJSONObject("file");
        if (file == null) {
            file = new JSONObject();
        }

        int fileId = file.optInt("id", 0);
        String name = file.optString("name", "");
        String format = file.optString("format", "");

        String title = "";
        String artist = "";
        String album = "";
        String genre = "";
        String year = "";
        int trackNumber = 0;
        double durationSeconds = 0.0;
        String metadataFormat = format;

        JSONObject metadata = file.optJSONObject("metadata");
        if (metadata != null) {
            title = metadata.optString("title", "");
            artist = metadata.optString("artist", "");
            album = metadata.optString("album", "");
            genre = metadata.optString("genre", "");
            year = metadata.optString("year", "");
            trackNumber = metadata.optInt("track_number", 0);
            durationSeconds = metadata.optDouble("length", 0.0);
            metadataFormat = metadata.optString("format", format);
        }

        return new Track(
                id,
                fileId,
                position,
                name,
                title,
                artist,
                album,
                genre,
                year,
                trackNumber,
                durationSeconds,
                metadataFormat,
                addedAt
        );
    }

    public static PaginatedResult<Track> trackListFromJson(JSONObject json) throws JSONException {
        JSONArray itemsArray = json.optJSONArray("items");
        List<Track> items = new ArrayList<Track>();
        if (itemsArray != null) {
            for (int i = 0; i < itemsArray.length(); i++) {
                items.add(trackFromJson(itemsArray.getJSONObject(i)));
            }
        }

        Pagination pagination = FileMapper.paginationFromJson(json.optJSONObject("pagination"));
        return new PaginatedResult<Track>(items, pagination);
    }

    public static MusicPlayerState playerStateFromJson(JSONObject json) throws JSONException {
        return new MusicPlayerState(
                json.optInt("id", 0),
                json.optString("client_id", ""),
                json.optInt("playlist_id", 0),
                json.optInt("current_file_id", 0),
                json.optDouble("current_position", 0.0),
                json.optDouble("volume", 1.0),
                json.optBoolean("shuffle", false),
                json.optString("repeat_mode", "none"),
                json.optString("updated_at", "")
        );
    }
}
