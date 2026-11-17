package com.kuranas.android.core.discovery

import android.content.Context
import android.net.wifi.WifiManager
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import java.net.InetAddress
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class UdpDiscovery @Inject constructor(@ApplicationContext private val context: Context) {

    /**
     * Envia o broadcast de descoberta e emite cada servidor KuraNAS que responder.
     * O fluxo termina após 3 segundos, liberando o multicast lock e o socket.
     */
    fun discover(): Flow<DiscoveredServer> {
        val wifiManager = context.applicationContext.getSystemService(Context.WIFI_SERVICE) as WifiManager
        val multicastLock = wifiManager.createMulticastLock(MULTICAST_LOCK_TAG).apply { setReferenceCounted(false) }
        return UdpBroadcastScan(
            acquireMulticastLock = { multicastLock.acquire() },
            releaseMulticastLock = { if (multicastLock.isHeld) multicastLock.release() },
            targetAddress = InetAddress.getByName(BROADCAST_ADDRESS),
            targetPort = DISCOVERY_PORT,
            listenMillis = LISTEN_MILLIS,
        ).discover()
    }

    private companion object {
        const val MULTICAST_LOCK_TAG = "kuranas-udp-discovery"
        const val BROADCAST_ADDRESS = "255.255.255.255"
        const val DISCOVERY_PORT = 19520
        const val LISTEN_MILLIS = 3000L
    }
}
