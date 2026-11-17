package com.kuranas.android.core.discovery

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOn
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.InetAddress
import java.net.SocketTimeoutException

internal class UdpBroadcastScan(
    private val acquireMulticastLock: () -> Unit,
    private val releaseMulticastLock: () -> Unit,
    private val targetAddress: InetAddress,
    private val targetPort: Int,
    private val listenMillis: Long,
) {

    fun discover(): Flow<DiscoveredServer> = flow {
        acquireMulticastLock()
        try {
            DatagramSocket().use { socket ->
                socket.broadcast = true
                sendProbe(socket)
                collectResponses(socket) { emit(it) }
            }
        } finally {
            releaseMulticastLock()
        }
    }.flowOn(Dispatchers.IO)

    private fun sendProbe(socket: DatagramSocket) {
        val probe = PROBE_MESSAGE.toByteArray(Charsets.UTF_8)
        socket.send(DatagramPacket(probe, probe.size, targetAddress, targetPort))
    }

    private suspend fun collectResponses(socket: DatagramSocket, onServer: suspend (DiscoveredServer) -> Unit) {
        val seenUrls = mutableSetOf<String>()
        val buffer = ByteArray(RESPONSE_BUFFER_SIZE)
        val deadline = System.currentTimeMillis() + listenMillis
        while (true) {
            val remainingMillis = deadline - System.currentTimeMillis()
            if (remainingMillis <= 0) return
            socket.soTimeout = remainingMillis.toInt()
            val packet = DatagramPacket(buffer, buffer.size)
            try {
                socket.receive(packet)
            } catch (_: SocketTimeoutException) {
                return
            }
            val payload = String(packet.data, 0, packet.length, Charsets.UTF_8)
            val server = UdpDiscoveryResponseParser.parse(payload, packet.address.hostAddress ?: continue)
            if (server != null && seenUrls.add(server.url)) onServer(server)
        }
    }

    companion object {
        const val PROBE_MESSAGE = "KURANAS_DISCOVER"
        private const val RESPONSE_BUFFER_SIZE = 1024
    }
}
