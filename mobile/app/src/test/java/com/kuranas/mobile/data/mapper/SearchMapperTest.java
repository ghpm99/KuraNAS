package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.SearchResult;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertTrue;

public class SearchMapperTest {

    @Test
    public void fromJson_withAllCategories_mapsCorrectly() throws JSONException {
        JSONObject fileObj = new JSONObject();
        fileObj.put("id", 1);
        fileObj.put("name", "document.pdf");
        fileObj.put("path", "/docs/document.pdf");
        fileObj.put("format", "pdf");

        JSONObject folderObj = new JSONObject();
        folderObj.put("id", 2);
        folderObj.put("name", "Photos");
        folderObj.put("path", "/Photos");

        JSONObject videoObj = new JSONObject();
        videoObj.put("id", 3);
        videoObj.put("name", "clip.mp4");
        videoObj.put("path", "/videos/clip.mp4");
        videoObj.put("format", "mp4");

        JSONObject imageObj = new JSONObject();
        imageObj.put("id", 4);
        imageObj.put("name", "photo.jpg");
        imageObj.put("path", "/images/photo.jpg");
        imageObj.put("format", "jpg");

        JSONArray files = new JSONArray();
        files.put(fileObj);

        JSONArray folders = new JSONArray();
        folders.put(folderObj);

        JSONArray videos = new JSONArray();
        videos.put(videoObj);

        JSONArray images = new JSONArray();
        images.put(imageObj);

        JSONObject json = new JSONObject();
        json.put("query", "test");
        json.put("suggestion", "testing");
        json.put("files", files);
        json.put("folders", folders);
        json.put("videos", videos);
        json.put("images", images);

        SearchResult result = SearchMapper.fromJson(json);

        assertEquals("test", result.getQuery());
        assertEquals("testing", result.getSuggestion());
        assertEquals(1, result.getFiles().size());
        assertEquals("document.pdf", result.getFiles().get(0).getName());
        assertTrue(result.getFiles().get(0).isFile());
        assertEquals(1, result.getFolders().size());
        assertEquals("Photos", result.getFolders().get(0).getName());
        assertTrue(result.getFolders().get(0).isDirectory());
        assertEquals(1, result.getVideos().size());
        assertEquals("clip.mp4", result.getVideos().get(0).getName());
        assertEquals(1, result.getImages().size());
        assertEquals("photo.jpg", result.getImages().get(0).getName());
        assertFalse(result.isEmpty());
        assertEquals(4, result.totalCount());
    }

    @Test
    public void fromJson_withEmptyCategories_returnsEmptyLists() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("query", "nothing");
        json.put("suggestion", "");
        json.put("files", new JSONArray());
        json.put("folders", new JSONArray());
        json.put("videos", new JSONArray());
        json.put("images", new JSONArray());

        SearchResult result = SearchMapper.fromJson(json);

        assertEquals("nothing", result.getQuery());
        assertTrue(result.getFiles().isEmpty());
        assertTrue(result.getFolders().isEmpty());
        assertTrue(result.getVideos().isEmpty());
        assertTrue(result.getImages().isEmpty());
        assertTrue(result.isEmpty());
        assertEquals(0, result.totalCount());
    }

    @Test
    public void fromJson_withMissingCategories_returnsEmptyLists() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("query", "search term");

        SearchResult result = SearchMapper.fromJson(json);

        assertNotNull(result);
        assertEquals("search term", result.getQuery());
        assertTrue(result.getFiles().isEmpty());
        assertTrue(result.getFolders().isEmpty());
        assertTrue(result.getVideos().isEmpty());
        assertTrue(result.getImages().isEmpty());
        assertTrue(result.isEmpty());
    }
}
