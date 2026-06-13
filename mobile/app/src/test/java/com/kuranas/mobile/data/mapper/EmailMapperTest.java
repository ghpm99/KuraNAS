package com.kuranas.mobile.data.mapper;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

import com.kuranas.mobile.domain.model.EmailItem;

import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

import java.util.List;

public class EmailMapperTest {

    @Test
    public void fromPaginatedJson_mapsAnalyzedAndUnanalyzed() throws JSONException {
        String json = "{\"items\":["
                + "{\"sender_name\":\"Banco\",\"sender_address\":\"b@x.com\",\"subject\":\"Fatura\","
                + "\"snippet\":\"corpo\",\"summary\":\"sua fatura venceu\",\"importance\":\"high\",\"verdict\":\"safe\","
                + "\"received_at\":\"2026-06-13T10:20:30Z\"},"
                + "{\"sender_name\":\"\",\"sender_address\":\"spam@x.com\",\"subject\":\"Ganhe\","
                + "\"snippet\":\"clique\",\"verdict\":\"suspicious\",\"received_at\":\"2026-06-13T09:00:00Z\"}"
                + "]}";

        List<EmailItem> items = EmailMapper.fromPaginatedJson(new JSONObject(json));

        assertEquals(2, items.size());

        EmailItem first = items.get(0);
        assertEquals("Banco", first.getSenderName());
        assertEquals("Fatura", first.getSubject());
        assertEquals("sua fatura venceu", first.getSummary());
        assertTrue(first.isHighImportance());
        assertFalse(first.isFlagged());

        EmailItem second = items.get(1);
        assertEquals("", second.getSummary());
        assertTrue(second.isFlagged());
        assertFalse(second.isHighImportance());
    }

    @Test
    public void fromPaginatedJson_missingItems_returnsEmpty() throws JSONException {
        assertTrue(EmailMapper.fromPaginatedJson(new JSONObject("{}")).isEmpty());
    }

    @Test
    public void isFlagged_trueForMalicious() {
        EmailItem item = EmailMapper.fromJson(jsonWithVerdict("malicious"));
        assertTrue(item.isFlagged());
    }

    @Test
    public void isFlagged_falseForSafe() {
        EmailItem item = EmailMapper.fromJson(jsonWithVerdict("safe"));
        assertFalse(item.isFlagged());
    }

    private static JSONObject jsonWithVerdict(String verdict) {
        JSONObject obj = new JSONObject();
        try {
            obj.put("verdict", verdict);
        } catch (JSONException e) {
            throw new RuntimeException(e);
        }
        return obj;
    }
}
