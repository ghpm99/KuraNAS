package com.kuranas.android.feature.cameraupload.ui

import android.Manifest

internal fun mediaPermissionsFor(includeVideos: Boolean): List<String> =
    if (includeVideos) listOf(Manifest.permission.READ_MEDIA_IMAGES, Manifest.permission.READ_MEDIA_VIDEO)
    else listOf(Manifest.permission.READ_MEDIA_IMAGES)

internal fun permissionsToRequestFor(includeVideos: Boolean): List<String> =
    mediaPermissionsFor(includeVideos) + Manifest.permission.POST_NOTIFICATIONS

internal fun hasMediaAccess(grantedPermissions: Set<String>, includeVideos: Boolean): Boolean =
    grantedPermissions.containsAll(mediaPermissionsFor(includeVideos))

internal fun hasNotificationAccess(grantedPermissions: Set<String>): Boolean =
    Manifest.permission.POST_NOTIFICATIONS in grantedPermissions
