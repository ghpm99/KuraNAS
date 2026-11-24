package com.kuranas.android.feature.cameraupload.ui

import android.Manifest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class CameraUploadPermissionsTest {

    private val images = Manifest.permission.READ_MEDIA_IMAGES
    private val video = Manifest.permission.READ_MEDIA_VIDEO
    private val notifications = Manifest.permission.POST_NOTIFICATIONS

    @Test
    fun `requests images and notifications when videos are excluded`() {
        assertEquals(listOf(images, notifications), permissionsToRequestFor(includeVideos = false))
    }

    @Test
    fun `requests video permission as well when videos are included`() {
        assertEquals(listOf(images, video, notifications), permissionsToRequestFor(includeVideos = true))
    }

    @Test
    fun `media access needs only images when videos are excluded`() {
        assertTrue(hasMediaAccess(setOf(images), includeVideos = false))
    }

    @Test
    fun `media access needs video permission too when videos are included`() {
        assertFalse(hasMediaAccess(setOf(images), includeVideos = true))
        assertTrue(hasMediaAccess(setOf(images, video), includeVideos = true))
    }

    @Test
    fun `no media access without the images permission`() {
        assertFalse(hasMediaAccess(setOf(video, notifications), includeVideos = false))
    }

    @Test
    fun `notification access is independent of media access`() {
        assertTrue(hasNotificationAccess(setOf(notifications)))
        assertFalse(hasNotificationAccess(setOf(images)))
    }
}
