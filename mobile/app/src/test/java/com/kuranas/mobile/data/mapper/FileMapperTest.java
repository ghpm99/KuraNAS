package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.Pagination;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertTrue;

public class FileMapperTest {

    @Test
    public void fromJson_withAllFields_mapsCorrectly() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("id", 42);
        json.put("name", "photo.jpg");
        json.put("path", "/photos/photo.jpg");
        json.put("parent_path", "/photos");
        json.put("type", FileItem.TYPE_FILE);
        json.put("format", "jpg");
        json.put("size", 102400L);
        json.put("updated_at", "2026-01-15T10:00:00Z");
        json.put("created_at", "2026-01-10T08:00:00Z");
        json.put("starred", true);
        json.put("directory_content_count", 0);

        FileItem item = FileMapper.fromJson(json);

        assertEquals(42, item.getId());
        assertEquals("photo.jpg", item.getName());
        assertEquals("/photos/photo.jpg", item.getPath());
        assertEquals("/photos", item.getParentPath());
        assertEquals(FileItem.TYPE_FILE, item.getType());
        assertEquals("jpg", item.getFormat());
        assertEquals(102400L, item.getSize());
        assertEquals("2026-01-15T10:00:00Z", item.getUpdatedAt());
        assertEquals("2026-01-10T08:00:00Z", item.getCreatedAt());
        assertTrue(item.isStarred());
        assertEquals(0, item.getDirectoryContentCount());
    }

    @Test
    public void fromJson_withMissingOptionalFields_usesDefaults() throws JSONException {
        JSONObject json = new JSONObject();

        FileItem item = FileMapper.fromJson(json);

        assertEquals(0, item.getId());
        assertEquals("", item.getName());
        assertEquals("", item.getPath());
        assertEquals("", item.getParentPath());
        assertEquals(0, item.getType());
        assertEquals("", item.getFormat());
        assertEquals(0L, item.getSize());
        assertEquals("", item.getUpdatedAt());
        assertEquals("", item.getCreatedAt());
        assertFalse(item.isStarred());
        assertEquals(0, item.getDirectoryContentCount());
    }

    @Test
    public void fromPaginatedJson_withItems_parsesList() throws JSONException {
        JSONObject file1 = new JSONObject();
        file1.put("id", 1);
        file1.put("name", "a.txt");
        file1.put("type", FileItem.TYPE_FILE);

        JSONObject file2 = new JSONObject();
        file2.put("id", 2);
        file2.put("name", "b.txt");
        file2.put("type", FileItem.TYPE_FILE);

        JSONArray items = new JSONArray();
        items.put(file1);
        items.put(file2);

        JSONObject pagination = new JSONObject();
        pagination.put("page", 1);
        pagination.put("page_size", 20);
        pagination.put("has_next", true);
        pagination.put("has_prev", false);

        JSONObject json = new JSONObject();
        json.put("items", items);
        json.put("pagination", pagination);

        PaginatedResult<FileItem> result = FileMapper.fromPaginatedJson(json);

        assertNotNull(result);
        assertEquals(2, result.getItems().size());
        assertEquals("a.txt", result.getItems().get(0).getName());
        assertEquals("b.txt", result.getItems().get(1).getName());
        assertTrue(result.hasNext());
        assertEquals(1, result.getPagination().getPage());
        assertEquals(20, result.getPagination().getPageSize());
        assertFalse(result.getPagination().hasPrev());
    }

    @Test
    public void fromPaginatedJson_withEmptyItems_returnsEmptyList() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("items", new JSONArray());

        PaginatedResult<FileItem> result = FileMapper.fromPaginatedJson(json);

        assertNotNull(result);
        assertTrue(result.getItems().isEmpty());
    }

    @Test
    public void fromPaginatedJson_withNullItems_returnsEmptyList() throws JSONException {
        JSONObject json = new JSONObject();

        PaginatedResult<FileItem> result = FileMapper.fromPaginatedJson(json);

        assertNotNull(result);
        assertTrue(result.getItems().isEmpty());
    }

    @Test
    public void paginationFromJson_withValidJson_mapsCorrectly() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("page", 3);
        json.put("page_size", 50);
        json.put("has_next", true);
        json.put("has_prev", true);

        Pagination pagination = FileMapper.paginationFromJson(json);

        assertEquals(3, pagination.getPage());
        assertEquals(50, pagination.getPageSize());
        assertTrue(pagination.hasNext());
        assertTrue(pagination.hasPrev());
    }

    @Test
    public void paginationFromJson_withNull_returnsDefaults() {
        Pagination pagination = FileMapper.paginationFromJson(null);

        assertEquals(1, pagination.getPage());
        assertEquals(20, pagination.getPageSize());
        assertFalse(pagination.hasNext());
        assertFalse(pagination.hasPrev());
    }
}
