package com.kuranas.android.feature.notifications.ui

import com.kuranas.android.feature.notifications.data.NotificationsApi
import com.kuranas.android.feature.notifications.data.UnreadCountDto
import io.mockk.coEvery
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import kotlinx.coroutines.withTimeout
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test
import java.io.IOException

@OptIn(ExperimentalCoroutinesApi::class)
class NotificationsIndicatorViewModelTest {

    private val api = mockk<NotificationsApi>()

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun NotificationsIndicatorViewModel.awaitUnreadCount(expected: Int) = runBlocking {
        withTimeout(AWAIT_TIMEOUT_MILLIS) { unreadCount.first { it == expected } }
    }

    @Test
    fun `loads the unread count on start`() {
        coEvery { api.getUnreadCount() } returns UnreadCountDto(count = 5)

        assertEquals(5, NotificationsIndicatorViewModel(api).awaitUnreadCount(5))
    }

    @Test
    fun `refresh picks up the new count`() {
        coEvery { api.getUnreadCount() } returns UnreadCountDto(count = 5)
        val viewModel = NotificationsIndicatorViewModel(api)
        viewModel.awaitUnreadCount(5)

        coEvery { api.getUnreadCount() } returns UnreadCountDto(count = 2)
        viewModel.refresh()

        assertEquals(2, viewModel.awaitUnreadCount(2))
    }

    @Test
    fun `a failing request clears the badge`() {
        coEvery { api.getUnreadCount() } returns UnreadCountDto(count = 5)
        val viewModel = NotificationsIndicatorViewModel(api)
        viewModel.awaitUnreadCount(5)

        coEvery { api.getUnreadCount() } throws IOException("offline")
        viewModel.refresh()

        assertEquals(0, viewModel.awaitUnreadCount(0))
    }

    private companion object {
        const val AWAIT_TIMEOUT_MILLIS = 5_000L
    }
}
