// File: components/live-preview/live-preview.mjs
import { loadCSS } from '../../shared/utils.mjs';
import { setPreviewStatus } from '../../shared/app-state.mjs';

class WHEPClient {
    constructor(videoElement, overlayElement, endpointUrl) {
        this.videoElement = videoElement; this.overlayElement = overlayElement; this.endpointUrl = endpointUrl;
        this.sessionUrl = ''; this.pc = null; this.state = 'idle';
        this.displayStream = new MediaStream(); this.realTracks = []; this.mockTracks = this.createMockTracks();
        // Store bound handlers for cleanup
        this.handlePlay = () => { if (this.state === 'idle') { this.videoElement.muted = false; this.connect(); } };
        this.handlePause = () => { if (this.state === 'connected' && !this.videoElement.ended) { this.disconnect('Paused by user'); } };
        this.initialize();
    }
    initialize() {
        this.mockTracks.forEach(track => this.displayStream.addTrack(track));
        this.videoElement.srcObject = this.displayStream; this.videoElement.muted = true;
        this.videoElement.addEventListener('play', this.handlePlay);
        this.videoElement.addEventListener('pause', this.handlePause);
    }
    createMockTracks() {
        const canvas = document.createElement('canvas'); canvas.width = 16; canvas.height = 9; const ctx = canvas.getContext('2d'); ctx.fillStyle = 'black'; ctx.fillRect(0, 0, 16, 9);
        const videoTrack = canvas.captureStream(1).getVideoTracks()[0]; videoTrack.enabled = true;
        // Store AudioContext for cleanup
        this.audioCtx = new (window.AudioContext || window.webkitAudioContext)(); const destination = this.audioCtx.createMediaStreamDestination();
        const audioTrack = destination.stream.getAudioTracks()[0]; audioTrack.enabled = true;
        return [videoTrack, audioTrack];
    }
    async connect() {
        if (this.state !== 'idle' && this.state !== 'reconnecting') return;
        this.state = 'connecting';
        this.showOverlay(true, 'Connecting...');
        setPreviewStatus('connecting');
        try {
            this.pc = new RTCPeerConnection({ iceServers: [{ urls: 'stun:stun.l.google.com:19302' }] });
            this.pc.ontrack = (event) => this.replaceTrack(event.track);
            this.pc.onconnectionstatechange = () => {
                if (!this.pc) return;
                if (this.pc.connectionState === 'connected') {
                    this.state = 'connected';
                    this.showOverlay(false);
                    setPreviewStatus('connected');
                    this.videoElement.play().catch(() => {});
                }
                else if (this.pc.connectionState === 'failed' && this.state !== 'reconnecting') { this.reconnect(); }
                else if (this.pc.connectionState === 'closed') { this.disconnect('Connection closed'); }
            };
            this.pc.addTransceiver('video', { direction: 'recvonly' });
            this.pc.addTransceiver('audio', { direction: 'recvonly' });
            const offer = await this.pc.createOffer();
            await this.pc.setLocalDescription(offer);
            const response = await fetch(this.endpointUrl, { method: 'POST', headers: { 'Content-Type': 'application/sdp' }, body: this.pc.localDescription.sdp });

            // Handle different HTTP error codes
            if (!response.ok) {
                if (response.status === 503) {
                    // Service Unavailable - no stream available (expected condition)
                    throw new Error('NO_STREAM_AVAILABLE');
                } else {
                    // Other errors
                    throw new Error(`Server responded with ${response.status}`);
                }
            }

            this.sessionUrl = response.headers.get('Location');
            const answerSDP = await response.text();
            await this.pc.setRemoteDescription({ type: 'answer', sdp: answerSDP });
        } catch (error) {
            // Handle specific error cases
            if (error.message === 'NO_STREAM_AVAILABLE') {
                // No stream available - show friendly message, don't log error
                this.disconnect('No stream');
                this.showOverlay(true, 'No Stream Available');
                setPreviewStatus('unavailable');
            } else {
                // Actual error - log to console
                console.error('WHEP Connection error:', error);
                this.disconnect(`Error: ${error.message}`);
                this.showOverlay(true, 'Connection Error');
                setPreviewStatus('unavailable');
                setTimeout(() => this.showOverlay(false), 2000);
            }
        }
    }
    replaceTrack(newTrack) {
        const mockTrack = this.mockTracks.find(t => t.kind === newTrack.kind); if (mockTrack) this.displayStream.removeTrack(mockTrack);
        this.displayStream.addTrack(newTrack); this.realTracks.push(newTrack);
        newTrack.onended = () => this.disconnect('Stream ended by remote');
    }
    disconnect(reason = 'Disconnect requested') {
        if (this.state === 'idle') return;
        if (this.sessionUrl) fetch(this.sessionUrl, { method: 'DELETE' }).catch(e => console.error("Error sending DELETE:", e));
        if (this.pc) { this.pc.close(); this.pc = null; } this.sessionUrl = '';
        this.realTracks.forEach(track => { track.stop(); this.displayStream.removeTrack(track); });
        this.realTracks = [];
        this.mockTracks.forEach(track => { if (!this.displayStream.getTracks().find(t => t.kind === track.kind)) this.displayStream.addTrack(track); });
        this.videoElement.muted = true; this.state = 'idle';
        if (reason !== 'Preparing for reconnect') {
            this.showOverlay(false);
            // Only set unavailable if stream ended remotely or actual error, not user pause
            if (reason !== 'Paused by user') {
                setPreviewStatus('unavailable');
            } else {
                setPreviewStatus('available');
            }
        }
    }
    reconnect() {
        if (this.state === 'reconnecting') return; this.state = 'reconnecting';
        this.showOverlay(true, 'Signal Lost. Reconnecting...');
        this.disconnect('Preparing for reconnect'); setTimeout(() => this.connect(), 2000);
    }
    showOverlay(show, text = 'Signal Lost') { this.overlayElement.innerText = text; this.overlayElement.style.display = show ? 'flex' : 'none'; }
    destroy() {
        // Disconnect active connection
        this.disconnect('Client destroyed');

        // Remove event listeners
        this.videoElement.removeEventListener('play', this.handlePlay);
        this.videoElement.removeEventListener('pause', this.handlePause);

        // Stop and cleanup mock tracks
        this.mockTracks.forEach(track => track.stop());
        this.mockTracks = [];

        // Close AudioContext
        if (this.audioCtx && this.audioCtx.state !== 'closed') {
            this.audioCtx.close();
            this.audioCtx = null;
        }

        // Cleanup video element
        if (this.videoElement.srcObject) {
            this.videoElement.srcObject.getTracks().forEach(track => track.stop());
            this.videoElement.srcObject = null;
        }

        // Clear display stream
        this.displayStream = null;

        console.log('WHEPClient destroyed and resources cleaned up');
    }
}

// Store active client instance for cleanup
let activeClient = null;

export function initLivePreview(container) {
    loadCSS('components/live-preview/live-preview.css');
    container.innerHTML = `
        <div class="video-wrapper">
            <video id="videoElement" controls></video>
            <div id="signalLostOverlay" class="overlay"></div>
        </div>
    `;
    const videoEl = container.querySelector('#videoElement');
    const overlayEl = container.querySelector('#signalLostOverlay');

    // Cleanup previous client if exists
    if (activeClient) {
        activeClient.destroy();
    }

    // Create new client and store reference
    activeClient = new WHEPClient(videoEl, overlayEl, '/whep/');
    return activeClient;
}

export function cleanupLivePreview() {
    if (activeClient) {
        activeClient.destroy();
        activeClient = null;
    }
}