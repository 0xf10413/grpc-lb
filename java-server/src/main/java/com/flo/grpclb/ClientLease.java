package com.flo.grpclb;

public class ClientLease {
    private final long actualStartTimestamp = System.currentTimeMillis();
    private long startTimestamp = actualStartTimestamp;
    private final long leaseDuration;
    private boolean forceExpired = false;

    public ClientLease(long leaseDuration) {
        this.leaseDuration = leaseDuration;
    }

    public void renew() {
        this.startTimestamp = System.currentTimeMillis();
    }

    /**
     * Marks lease as expired.
     * It's not possible to renew it anymore after.
     */
    public void expire() {
        forceExpired = true;
    }

    public boolean expired() {
        return forceExpired || this.startTimestamp + leaseDuration < System.currentTimeMillis();
    }

    public boolean forceExpired() {
        return forceExpired;
    }

    public long getActualStartTimestamp() {
        return actualStartTimestamp;
    }
}
