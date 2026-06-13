package com.kuranas.mobile.data.mapper;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertTrue;

import com.kuranas.mobile.domain.model.NotificationItem;

import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

import java.util.List;

public class NotificationMapperTest {

    @Test
    public void fromPaginatedJson_mapsItems() throws JSONException {
        String json = "{\"items\":["
                + "{\"id\":1,\"type\":\"warning\",\"title\":\"Disco cheio\",\"message\":\"95%\",\"created_at\":\"2026-06-13T10:20:30Z\"},"
                + "{\"id\":2,\"type\":\"info\",\"title\":\"Scan\",\"message\":\"ok\",\"created_at\":\"2026-06-13T11:00:00Z\"}"
                + "],\"pagination\":{\"page\":1}}";

        List<NotificationItem> items = NotificationMapper.fromPaginatedJson(new JSONObject(json));

        assertEquals(2, items.size());
        assertEquals("warning", items.get(0).getType());
        assertEquals("Disco cheio", items.get(0).getTitle());
        assertEquals("95%", items.get(0).getMessage());
        assertEquals("2026-06-13T10:20:30Z", items.get(0).getCreatedAt());
    }

    @Test
    public void fromPaginatedJson_missingItems_returnsEmpty() throws JSONException {
        List<NotificationItem> items = NotificationMapper.fromPaginatedJson(new JSONObject("{\"pagination\":{}}"));
        assertTrue(items.isEmpty());
    }

    @Test
    public void fromPaginatedJson_nullObject_returnsEmpty() throws JSONException {
        assertTrue(NotificationMapper.fromPaginatedJson(null).isEmpty());
    }

    @Test
    public void fromJson_missingFields_defaultToEmpty() {
        NotificationItem item = NotificationMapper.fromJson(new JSONObject());
        assertEquals("", item.getType());
        assertEquals("", item.getTitle());
        assertEquals("", item.getMessage());
        assertEquals("", item.getCreatedAt());
    }
}
