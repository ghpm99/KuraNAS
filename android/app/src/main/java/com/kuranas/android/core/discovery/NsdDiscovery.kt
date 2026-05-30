package com.kuranas.android.core.discovery

import android.content.Context
import android.net.nsd.NsdManager
import android.net.nsd.NsdServiceInfo
import android.util.Log
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import javax.inject.Inject
import javax.inject.Singleton

data class DiscoveredServer(val name: String, val host: String, val port: Int) {
    val url: String get() = "http://$host:$port"
}

@Singleton
class NsdDiscovery @Inject constructor(@ApplicationContext private val context: Context) {

    private val nsdManager = context.getSystemService(Context.NSD_SERVICE) as NsdManager
    private val serviceType = "_nas._tcp."

    fun discover(): Flow<DiscoveredServer> = callbackFlow {
        val resolvedServers = mutableSetOf<String>()

        val discoveryListener = object : NsdManager.DiscoveryListener {
            override fun onStartDiscoveryFailed(serviceType: String, errorCode: Int) {
                Log.w("NsdDiscovery", "Start failed: $errorCode")
            }
            override fun onStopDiscoveryFailed(serviceType: String, errorCode: Int) {
                Log.w("NsdDiscovery", "Stop failed: $errorCode")
            }
            override fun onDiscoveryStarted(serviceType: String) {
                Log.d("NsdDiscovery", "Discovery started")
            }
            override fun onDiscoveryStopped(serviceType: String) {
                Log.d("NsdDiscovery", "Discovery stopped")
            }
            override fun onServiceLost(serviceInfo: NsdServiceInfo) {}
            override fun onServiceFound(serviceInfo: NsdServiceInfo) {
                if (resolvedServers.contains(serviceInfo.serviceName)) return
                nsdManager.resolveService(serviceInfo, object : NsdManager.ResolveListener {
                    override fun onResolveFailed(si: NsdServiceInfo, errorCode: Int) {
                        Log.w("NsdDiscovery", "Resolve failed for ${si.serviceName}: $errorCode")
                    }
                    override fun onServiceResolved(si: NsdServiceInfo) {
                        val host = si.host?.hostAddress ?: return
                        val server = DiscoveredServer(si.serviceName, host, si.port)
                        resolvedServers.add(si.serviceName)
                        trySend(server)
                    }
                })
            }
        }

        try {
            nsdManager.discoverServices(serviceType, NsdManager.PROTOCOL_DNS_SD, discoveryListener)
        } catch (e: Exception) {
            Log.e("NsdDiscovery", "Failed to start discovery", e)
        }

        awaitClose {
            try { nsdManager.stopServiceDiscovery(discoveryListener) } catch (_: Exception) {}
        }
    }
}
