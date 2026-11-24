package com.kuranas.android.feature.cameraupload.data

import kotlinx.serialization.json.Json
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class ServerReachabilityTest {

    private val json = Json { ignoreUnknownKeys = true }

    @Test
    fun healthyOnSuccessWithStatusOk() {
        assertTrue(isHealthyKuranasResponse(200, """{"service":"kuranas","status":"ok"}""", json))
    }

    @Test
    fun notHealthyOnErrorStatusCode() {
        assertFalse(isHealthyKuranasResponse(503, """{"status":"ok"}""", json))
    }

    @Test
    fun notHealthyWhenStatusIsNotOk() {
        assertFalse(isHealthyKuranasResponse(200, """{"status":"degraded"}""", json))
    }

    @Test
    fun notHealthyOnMissingOrInvalidBody() {
        assertFalse(isHealthyKuranasResponse(200, null, json))
        assertFalse(isHealthyKuranasResponse(200, "<html>router</html>", json))
    }
}
