package com.kuranas.android.core.discovery

import kotlinx.coroutines.flow.toList
import kotlinx.coroutines.runBlocking
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.InetAddress
import kotlin.concurrent.thread

class UdpBroadcastScanTest {

    private val loopback = InetAddress.getByName("127.0.0.1")

    private fun startFakeServer(responsePayloads: List<String>): Pair<DatagramSocket, Thread> {
        val serverSocket = DatagramSocket(0, loopback)
        val worker = thread {
            val buffer = ByteArray(64)
            val probe = DatagramPacket(buffer, buffer.size)
            serverSocket.receive(probe)
            responsePayloads.forEach { payload ->
                val bytes = payload.toByteArray()
                serverSocket.send(DatagramPacket(bytes, bytes.size, probe.address, probe.port))
            }
        }
        return serverSocket to worker
    }

    @Test
    fun `emits deduplicated servers, ends after listen window and releases lock`() = runBlocking {
        val validResponse = """{"service":"kuranas","port":8000,"api":"/api/v1"}"""
        val (serverSocket, worker) = startFakeServer(listOf(validResponse, validResponse, """{"service":"x","port":1}"""))
        val lockEvents = mutableListOf<String>()

        val scan = UdpBroadcastScan(
            acquireMulticastLock = { lockEvents.add("acquire") },
            releaseMulticastLock = { lockEvents.add("release") },
            targetAddress = loopback,
            targetPort = serverSocket.localPort,
            listenMillis = 600,
        )
        val startedAt = System.currentTimeMillis()
        val servers = scan.discover().toList()
        val elapsedMillis = System.currentTimeMillis() - startedAt
        worker.join()
        serverSocket.close()

        assertEquals(listOf("http://127.0.0.1:8000"), servers.map { it.url })
        assertEquals(listOf("acquire", "release"), lockEvents)
        assertTrue("elapsed=$elapsedMillis", elapsedMillis in 500..2000)
    }
}
