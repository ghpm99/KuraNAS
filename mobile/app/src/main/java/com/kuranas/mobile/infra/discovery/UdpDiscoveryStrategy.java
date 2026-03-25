package com.kuranas.mobile.infra.discovery;

import android.util.Log;

import org.json.JSONObject;

import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;

public final class UdpDiscoveryStrategy implements DiscoveryStrategy {

    private static final String TAG = "UdpDiscovery";
    private static final String DISCOVERY_MESSAGE = "KURANAS_DISCOVER";
    private static final int DISCOVERY_PORT = 19520;
    private static final int RECEIVE_TIMEOUT_MS = 3000;

    private final ServerValidator validator;
    private volatile DatagramSocket socket;
    private volatile boolean cancelled;

    public UdpDiscoveryStrategy(ServerValidator validator) {
        this.validator = validator;
    }

    @Override
    public String name() {
        return "udp";
    }

    @Override
    public void discover(StrategyCallback callback) {
        cancelled = false;
        DatagramSocket localSocket = null;

        try {
            localSocket = new DatagramSocket();
            socket = localSocket;
            localSocket.setBroadcast(true);
            localSocket.setSoTimeout(RECEIVE_TIMEOUT_MS);

            byte[] sendData = DISCOVERY_MESSAGE.getBytes("UTF-8");
            DatagramPacket sendPacket = new DatagramPacket(
                    sendData,
                    sendData.length,
                    InetAddress.getByName("255.255.255.255"),
                    DISCOVERY_PORT
            );

            localSocket.send(sendPacket);
            Log.d(TAG, "Sent broadcast discovery packet");

            byte[] receiveBuffer = new byte[1024];
            DatagramPacket receivePacket = new DatagramPacket(receiveBuffer, receiveBuffer.length);

            localSocket.receive(receivePacket);

            if (cancelled) {
                callback.onNotFound();
                return;
            }

            String response = new String(receivePacket.getData(), 0, receivePacket.getLength(), "UTF-8");
            Log.d(TAG, "Received response: " + response);

            JSONObject json = new JSONObject(response);
            String service = json.optString("service", "");
            int port = json.optInt("port", 0);

            if (!"kuranas".equals(service) || port == 0) {
                callback.onNotFound();
                return;
            }

            String senderHost = receivePacket.getAddress().getHostAddress();
            String candidateUrl = "http://" + senderHost + ":" + port;

            if (validator.validate(candidateUrl)) {
                callback.onFound(candidateUrl);
            } else {
                callback.onNotFound();
            }

        } catch (Exception e) {
            Log.d(TAG, "UDP discovery failed: " + e.getMessage());
            callback.onNotFound();
        } finally {
            if (localSocket != null && !localSocket.isClosed()) {
                localSocket.close();
            }
            socket = null;
        }
    }

    @Override
    public void cancel() {
        cancelled = true;
        DatagramSocket s = socket;
        if (s != null && !s.isClosed()) {
            s.close();
        }
    }
}
