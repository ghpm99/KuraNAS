package com.kuranas.android.feature.files.data

import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.data.TransferEntity
import com.kuranas.android.feature.transfers.data.TransferStatus
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class NewUploadCompletionDetectorTest {

    private fun transfer(
        id: Long,
        status: TransferStatus,
        folderId: Long? = 5,
        direction: TransferDirection = TransferDirection.UPLOAD,
    ) = TransferEntity(
        id = id,
        batchId = "batch",
        direction = direction,
        displayName = "f$id",
        targetFolderId = folderId,
        targetFolderName = "folder",
        status = status,
        createdAt = 0,
        updatedAt = 0,
    )

    private fun detectorWithBaseline(vararg baseline: TransferEntity) =
        NewUploadCompletionDetector().also { it.hasNewCompletionIn(baseline.toList(), 5) }

    @Test
    fun `completion after baseline in the open folder is new`() {
        val detector = detectorWithBaseline(transfer(1, TransferStatus.RUNNING))

        assertTrue(detector.hasNewCompletionIn(listOf(transfer(1, TransferStatus.COMPLETED)), 5))
    }

    @Test
    fun `completed rows present at the baseline do not count`() {
        val detector = NewUploadCompletionDetector()

        assertFalse(detector.hasNewCompletionIn(listOf(transfer(1, TransferStatus.COMPLETED)), 5))
        assertFalse(detector.hasNewCompletionIn(listOf(transfer(1, TransferStatus.COMPLETED)), 5))
    }

    @Test
    fun `a completion is reported only once`() {
        val detector = detectorWithBaseline()
        val completed = listOf(transfer(1, TransferStatus.COMPLETED))

        assertTrue(detector.hasNewCompletionIn(completed, 5))
        assertFalse(detector.hasNewCompletionIn(completed, 5))
    }

    @Test
    fun `completion targeting another folder does not count nor resurfaces after navigating there`() {
        val detector = detectorWithBaseline()
        val completedElsewhere = listOf(transfer(1, TransferStatus.COMPLETED, folderId = 9))

        assertFalse(detector.hasNewCompletionIn(completedElsewhere, 5))
        assertFalse(detector.hasNewCompletionIn(completedElsewhere, 9))
    }

    @Test
    fun `root is matched by a null target folder`() {
        val detector = detectorWithBaseline()

        assertTrue(detector.hasNewCompletionIn(listOf(transfer(1, TransferStatus.COMPLETED, folderId = null)), null))
    }

    @Test
    fun `already existed and failed uploads need no refresh`() {
        val detector = detectorWithBaseline()

        assertFalse(
            detector.hasNewCompletionIn(
                listOf(transfer(1, TransferStatus.ALREADY_EXISTED), transfer(2, TransferStatus.FAILED)),
                5,
            ),
        )
    }

    @Test
    fun `completed downloads are ignored`() {
        val detector = detectorWithBaseline()

        assertFalse(
            detector.hasNewCompletionIn(
                listOf(transfer(1, TransferStatus.COMPLETED, direction = TransferDirection.DOWNLOAD)),
                5,
            ),
        )
    }
}
