// ============================================
// VIBE COMPOSER — Frontend Application
// ============================================

(function () {
    'use strict';

    // --- State ---
    let authHeader = '';
    let currentUsername = '';
    let pollingInterval = null;
    let activeCompositionId = null;

    // Recording state (숙성)
    let isRecording = false;
    let mediaRecorder = null;
    let recordingChunks = [];
    let recordingStartTime = null;
    let recordingTimerInterval = null;
    let clips = []; // loaded from server
    const MAX_RECORDING_DURATION_MS = 5 * 60 * 1000; // 5 minutes
    let maxDurationTimeout = null;

    // --- DOM Refs ---
    const $ = (sel) => document.querySelector(sel);
    const $$ = (sel) => document.querySelectorAll(sel);

    // Screens
    const loginScreen = $('#login-screen');
    const appScreen = $('#app-screen');

    // Login
    const loginForm = $('#login-form');
    const usernameInput = $('#username-input');
    const passwordInput = $('#password-input');
    const loginError = $('#login-error');

    // App
    const userBadge = $('#user-badge');
    const logoutBtn = $('#logout-btn');

    // Tabs
    const tabBtns = $$('.tab-btn');
    const composeTab = $('#compose-tab');
    const recordTab = $('#record-tab');
    const libraryTab = $('#library-tab');

    // Compose
    const ventText = $('#vent-text');
    const styleFunnyLabel = $('#style-funny-label');
    const styleHarshLabel = $('#style-harsh-label');
    const styleHiphopLabel = $('#style-hiphop-label');
    const stylePansoriLabel = $('#style-pansori-label');
    const voiceAnyLabel = $('#voice-any-label');
    const voiceMaleLabel = $('#voice-male-label');
    const voiceFemaleLabel = $('#voice-female-label');
    const generateBtn = $('#generate-btn');
    const composeError = $('#compose-error');
    const activeGeneration = $('#active-generation');
    const statusMessage = $('#status-message');

    // Lyric type
    const lyricArcLabel = $('#lyric-arc-label');
    const lyricImmersionLabel = $('#lyric-immersion-label');

    // Record (숙성)
    const recordBtn = $('#record-btn');
    const recordIcon = $('#record-icon');
    const recordTimer = $('#record-timer');
    const recordLabel = $('#record-label');
    const audioLevelBars = $('#audio-level-bars');
    const clipCount = $('#clip-count');
    const clipTotalDuration = $('#clip-total-duration');
    const clipList = $('#clip-list');
    const generateFromClipsBtn = $('#generate-from-clips-btn');
    const recordError = $('#record-error');

    // Library
    const compositionsList = $('#compositions-list');
    const refreshBtn = $('#refresh-btn');

    // Player
    const audioPlayer = $('#audio-player');
    const playerAudio = $('#player-audio');
    const playerTitle = $('#player-title');
    const playerCloseBtn = $('#player-close-btn');

    // --- Init ---
    function init() {
        // Restore session
        const saved = localStorage.getItem('vibe_auth');
        if (saved) {
            try {
                const data = JSON.parse(saved);
                authHeader = data.authHeader;
                currentUsername = data.username;
                verifyAuth();
            } catch {
                showLogin();
            }
        } else {
            showLogin();
        }

        bindEvents();
    }

    // --- Auth ---
    async function verifyAuth() {
        try {
            const res = await api('/api/me');
            if (res.ok) {
                const data = await res.json();
                currentUsername = data.username;
                showApp();
            } else {
                showLogin();
            }
        } catch {
            showLogin();
        }
    }

    function showLogin() {
        loginScreen.classList.add('active');
        appScreen.classList.remove('active');
        localStorage.removeItem('vibe_auth');
        authHeader = '';
        currentUsername = '';
    }

    function showApp() {
        loginScreen.classList.remove('active');
        appScreen.classList.add('active');
        userBadge.textContent = `@${currentUsername}`;
        loadCompositions();
        checkActiveGeneration();
    }

    async function handleLogin(e) {
        e.preventDefault();
        const username = usernameInput.value.trim();
        const password = passwordInput.value;

        if (!username) return;

        authHeader = 'Basic ' + btoa(username + ':' + password);
        loginError.classList.add('hidden');

        try {
            const res = await api('/api/me');
            if (res.ok) {
                currentUsername = username;
                localStorage.setItem('vibe_auth', JSON.stringify({
                    authHeader,
                    username: currentUsername,
                }));
                showApp();
            } else {
                const errData = await res.text();
                loginError.textContent = res.status === 403
                    ? 'Sorry, you\'re not invited 😔'
                    : 'Invalid credentials';
                loginError.classList.remove('hidden');
                authHeader = '';
            }
        } catch (err) {
            loginError.textContent = 'Connection error. Is the server running?';
            loginError.classList.remove('hidden');
            authHeader = '';
        }
    }

    function handleLogout() {
        stopPolling();
        stopRecording();
        showLogin();
    }

    // --- API Helper ---
    function api(path, options = {}) {
        const headers = { ...options.headers };
        if (authHeader) {
            headers['Authorization'] = authHeader;
        }
        return fetch(path, { ...options, headers });
    }

    // --- Tab Navigation ---
    function switchTab(tabName) {
        tabBtns.forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tab === tabName);
        });
        composeTab.classList.toggle('active', tabName === 'compose');
        recordTab.classList.toggle('active', tabName === 'record');
        libraryTab.classList.toggle('active', tabName === 'library');

        if (tabName === 'library') {
            loadCompositions();
        }
        if (tabName === 'record') {
            loadClips();
        }
    }



    // --- Style Selection ---
    function updateStyleSelection() {
        const selected = document.querySelector('input[name="style"]:checked').value;
        styleFunnyLabel.classList.toggle('selected', selected === 'funny');
        styleHarshLabel.classList.toggle('selected', selected === 'harsh');
        styleHiphopLabel.classList.toggle('selected', selected === 'hiphop');
        stylePansoriLabel.classList.toggle('selected', selected === 'pansori');
    }

    // --- Voice Selection ---
    function updateVoiceSelection() {
        const selected = document.querySelector('input[name="voice"]:checked').value;
        voiceAnyLabel.classList.toggle('selected', selected === 'any');
        voiceMaleLabel.classList.toggle('selected', selected === 'male');
        voiceFemaleLabel.classList.toggle('selected', selected === 'female');
    }

    // --- Lyric Type Selection ---
    function updateLyricTypeSelection() {
        const selected = document.querySelector('input[name="lyric_type"]:checked').value;
        lyricArcLabel.classList.toggle('selected', selected === 'arc');
        lyricImmersionLabel.classList.toggle('selected', selected === 'immersion');
    }

    // --- Record Tab: Style/Voice/Lyric Selection (separate radio group) ---
    function updateRecOptionSelection(groupName) {
        const selected = document.querySelector(`input[name="${groupName}"]:checked`);
        if (!selected) return;
        const options = document.querySelectorAll(`input[name="${groupName}"]`);
        options.forEach(input => {
            const label = input.closest('.style-option');
            if (label) {
                label.classList.toggle('selected', input === selected);
            }
        });
    }

    // --- Compose ---
    async function handleCompose() {
        const style = document.querySelector('input[name="style"]:checked').value;
        const text = ventText.value.trim();

        if (!text) {
            showComposeError('Please tell us what\'s bothering you!');
            return;
        }

        // Show loading state
        setGenerateLoading(true);
        composeError.classList.add('hidden');

        const formData = new FormData();
        formData.append('style', style);
        formData.append('voice', document.querySelector('input[name="voice"]:checked').value);
        formData.append('lyric_type', document.querySelector('input[name="lyric_type"]:checked').value);
        formData.append('text', text);

        try {
            const res = await api('/api/compose', {
                method: 'POST',
                body: formData,
            });

            const data = await res.json();

            if (res.ok) {
                activeCompositionId = data.id;
                ventText.value = '';
                showActiveGeneration(data.id);
                startPolling(data.id);
            } else if (res.status === 409) {
                showComposeError(data.error || 'You already have a composition in progress!');
            } else {
                showComposeError(data.error || 'Something went wrong');
            }
        } catch (err) {
            showComposeError('Connection error: ' + err.message);
        } finally {
            setGenerateLoading(false);
        }
    }

    function setGenerateLoading(loading) {
        generateBtn.disabled = loading;
        generateBtn.querySelector('.btn-text').classList.toggle('hidden', loading);
        generateBtn.querySelector('.btn-loading').classList.toggle('hidden', !loading);
    }

    function showComposeError(msg) {
        composeError.textContent = msg;
        composeError.classList.remove('hidden');
    }

    function showActiveGeneration() {
        activeGeneration.classList.remove('hidden');
    }

    function hideActiveGeneration() {
        activeGeneration.classList.add('hidden');
    }

    // --- Polling ---
    function startPolling(compositionId) {
        stopPolling();
        let step = 0;
        const messages = [
            'Analyzing your emotions...',
            'Detecting vocal tone and intensity...',
            'Crafting the perfect musical revenge...',
            'Lyria is composing your track...',
            'Adding the finishing touches...',
            'Almost there... this is going to be good 🔥',
        ];

        pollingInterval = setInterval(async () => {
            step = Math.min(step + 1, messages.length - 1);
            statusMessage.textContent = messages[step];

            try {
                const res = await api(`/api/compositions/${compositionId}`);
                if (!res.ok) return;

                const data = await res.json();

                if (data.status === 'done') {
                    stopPolling();
                    hideActiveGeneration();
                    activeCompositionId = null;
                    // Switch to library and reload
                    switchTab('library');
                    showNotification('Your music is ready! 🎵');
                } else if (data.status === 'failed') {
                    stopPolling();
                    hideActiveGeneration();
                    activeCompositionId = null;
                    showComposeError(data.error_message || 'Generation failed. Please try again.');
                }
            } catch {
                // Network error, keep polling
            }
        }, 5000);
    }

    function stopPolling() {
        if (pollingInterval) {
            clearInterval(pollingInterval);
            pollingInterval = null;
        }
    }

    async function checkActiveGeneration() {
        try {
            const res = await api('/api/compositions');
            if (!res.ok) return;

            const data = await res.json();
            const active = data.find(c => c.status === 'pending' || c.status === 'generating');
            if (active) {
                activeCompositionId = active.id;
                showActiveGeneration();
                startPolling(active.id);
            }
        } catch {
            // ignore
        }
    }

    // ============================================
    // 숙성 — RECORDING
    // ============================================

    async function toggleRecording() {
        if (isRecording) {
            stopRecording();
        } else {
            await startRecording();
        }
    }

    async function startRecording() {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            recordingChunks = [];

            mediaRecorder = new MediaRecorder(stream, {
                mimeType: MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
                    ? 'audio/webm;codecs=opus'
                    : 'audio/webm',
            });

            mediaRecorder.ondataavailable = (e) => {
                if (e.data.size > 0) {
                    recordingChunks.push(e.data);
                }
            };

            mediaRecorder.onstop = () => {
                const blob = new Blob(recordingChunks, { type: mediaRecorder.mimeType });
                const durationMs = Date.now() - recordingStartTime;
                stream.getTracks().forEach(t => t.stop());
                uploadClip(blob, durationMs);
            };

            mediaRecorder.start(1000); // collect in 1s chunks
            isRecording = true;
            recordingStartTime = Date.now();

            // UI updates
            recordBtn.classList.add('recording');
            recordIcon.textContent = '⏹';
            recordLabel.textContent = 'Recording... tap to stop';
            audioLevelBars.classList.remove('hidden');
            startRecordingTimer();

            // Auto-stop after 5 minutes
            maxDurationTimeout = setTimeout(() => {
                if (isRecording) {
                    stopRecording();
                    showNotification('Recording auto-stopped (5 min max) ⏱️');
                }
            }, MAX_RECORDING_DURATION_MS);

        } catch (err) {
            showRecordError('Microphone access denied. Please allow mic access.');
            console.error('Mic error:', err);
        }
    }

    function stopRecording() {
        if (!isRecording || !mediaRecorder) return;

        if (maxDurationTimeout) {
            clearTimeout(maxDurationTimeout);
            maxDurationTimeout = null;
        }

        mediaRecorder.stop();
        isRecording = false;

        // UI updates
        recordBtn.classList.remove('recording');
        recordIcon.textContent = '🎤';
        recordLabel.textContent = 'Tap to record';
        audioLevelBars.classList.add('hidden');
        stopRecordingTimer();
    }

    function startRecordingTimer() {
        recordTimer.textContent = '0:00';
        recordingTimerInterval = setInterval(() => {
            const elapsed = Date.now() - recordingStartTime;
            recordTimer.textContent = formatDuration(elapsed);
        }, 100);
    }

    function stopRecordingTimer() {
        if (recordingTimerInterval) {
            clearInterval(recordingTimerInterval);
            recordingTimerInterval = null;
        }
        recordTimer.textContent = '0:00';
    }

    async function uploadClip(blob, durationMs) {
        // Show uploading state in clip list
        const tempId = 'uploading-' + Date.now();
        addTempClipCard(tempId, durationMs);

        const formData = new FormData();
        formData.append('audio', blob, `clip-${Date.now()}.webm`);
        formData.append('duration_ms', String(Math.round(durationMs)));

        try {
            const res = await api('/api/clips', {
                method: 'POST',
                body: formData,
            });

            if (res.ok) {
                const data = await res.json();
                // Reload clips from server to get the full list
                await loadClips();
                showNotification('Clip saved ✓');
            } else {
                const errData = await res.json().catch(() => ({}));
                removeTempClipCard(tempId);
                showRecordError(errData.error || 'Failed to upload clip');
            }
        } catch (err) {
            removeTempClipCard(tempId);
            showRecordError('Upload failed: ' + err.message);
        }
    }

    async function loadClips() {
        try {
            const res = await api('/api/clips');
            if (!res.ok) return;

            clips = await res.json();
            renderClips();
            updateClipSummary();
            updateGenerateFromClipsBtn();
        } catch {
            // ignore
        }
    }

    function renderClips() {
        if (!clips || clips.length === 0) {
            clipList.innerHTML = `
                <div class="empty-state clip-empty">
                    <span class="empty-icon">🎙️</span>
                    <p>No recordings yet. Hit the mic to start!</p>
                </div>
            `;
            return;
        }

        clipList.innerHTML = clips.map((clip, i) => {
            const date = new Date(clip.created_at).toLocaleTimeString('en-US', {
                hour: '2-digit',
                minute: '2-digit',
            });
            return `
                <div class="clip-card" id="clip-${clip.id}">
                    <div class="clip-index">${i + 1}</div>
                    <div class="clip-info">
                        <div class="clip-duration">${formatDuration(clip.duration_ms)}</div>
                        <div class="clip-time">${date}</div>
                    </div>
                    <div class="clip-actions">
                        <button class="clip-play-btn" onclick="window.vibeApp.playClip('${clip.id}')" title="Play">▶</button>
                        <button class="clip-delete-btn" onclick="window.vibeApp.deleteClip('${clip.id}')" title="Delete">✕</button>
                    </div>
                </div>
            `;
        }).join('');
    }

    function addTempClipCard(tempId, durationMs) {
        // Remove empty state if present
        const empty = clipList.querySelector('.clip-empty');
        if (empty) empty.remove();

        const card = document.createElement('div');
        card.className = 'clip-card';
        card.id = tempId;
        card.innerHTML = `
            <div class="clip-index">•</div>
            <div class="clip-info">
                <div class="clip-duration">${formatDuration(durationMs)}</div>
                <div class="clip-time">Uploading...</div>
            </div>
            <div class="clip-actions">
                <div class="clip-uploading">
                    <span class="spinner"></span>
                </div>
            </div>
        `;
        clipList.appendChild(card);
    }

    function removeTempClipCard(tempId) {
        const el = document.getElementById(tempId);
        if (el) el.remove();
    }

    function updateClipSummary() {
        const count = clips ? clips.length : 0;
        const totalMs = clips ? clips.reduce((sum, c) => sum + c.duration_ms, 0) : 0;
        clipCount.textContent = `${count} clip${count !== 1 ? 's' : ''}`;
        clipTotalDuration.textContent = `${formatDuration(totalMs)} total`;
    }

    function updateGenerateFromClipsBtn() {
        const hasClips = clips && clips.length > 0;
        generateFromClipsBtn.disabled = !hasClips;
    }

    async function playClip(clipId) {
        try {
            const res = await api(`/api/clips/${clipId}/download`);
            if (!res.ok) throw new Error('Failed to load clip');
            const blob = await res.blob();
            const url = URL.createObjectURL(blob);
            playerAudio.src = url;
            playerTitle.textContent = `Clip #${getClipIndex(clipId)}`;
            audioPlayer.classList.remove('hidden');
            playerAudio.play();
        } catch (err) {
            showRecordError('Error playing clip: ' + err.message);
        }
    }

    async function deleteClip(clipId) {
        try {
            const res = await api(`/api/clips/${clipId}`, { method: 'DELETE' });
            if (res.ok) {
                await loadClips();
                showNotification('Clip deleted');
            } else {
                const errData = await res.json().catch(() => ({}));
                showRecordError(errData.error || 'Failed to delete clip');
            }
        } catch (err) {
            showRecordError('Delete failed: ' + err.message);
        }
    }

    function getClipIndex(clipId) {
        if (!clips) return '?';
        const idx = clips.findIndex(c => c.id === clipId);
        return idx >= 0 ? idx + 1 : '?';
    }

    // --- Generate From Clips ---
    async function handleGenerateFromClips() {
        if (!clips || clips.length === 0) {
            showRecordError('Record some clips first!');
            return;
        }

        const style = document.querySelector('input[name="rec_style"]:checked').value;
        const voice = document.querySelector('input[name="rec_voice"]:checked').value;
        const lyricType = document.querySelector('input[name="rec_lyric_type"]:checked').value;

        // Show loading
        generateFromClipsBtn.disabled = true;
        generateFromClipsBtn.querySelector('.btn-text').classList.add('hidden');
        generateFromClipsBtn.querySelector('.btn-loading').classList.remove('hidden');
        recordError.classList.add('hidden');

        const formData = new FormData();
        formData.append('source', 'clips');
        formData.append('style', style);
        formData.append('voice', voice);
        formData.append('lyric_type', lyricType);

        try {
            const res = await api('/api/compose', {
                method: 'POST',
                body: formData,
            });

            const data = await res.json();

            if (res.ok) {
                activeCompositionId = data.id;
                // Switch to compose tab to show progress
                switchTab('compose');
                showActiveGeneration();
                startPolling(data.id);
                showNotification(`Brewing ${clips.length} clips into music... 🍶`);
            } else if (res.status === 409) {
                showRecordError(data.error || 'You already have a composition in progress!');
            } else {
                showRecordError(data.error || 'Something went wrong');
            }
        } catch (err) {
            showRecordError('Connection error: ' + err.message);
        } finally {
            generateFromClipsBtn.disabled = false;
            generateFromClipsBtn.querySelector('.btn-text').classList.remove('hidden');
            generateFromClipsBtn.querySelector('.btn-loading').classList.add('hidden');
            updateGenerateFromClipsBtn();
        }
    }

    function showRecordError(msg) {
        recordError.textContent = msg;
        recordError.classList.remove('hidden');
    }

    // --- Library ---
    async function loadCompositions() {
        try {
            const res = await api('/api/compositions');
            if (!res.ok) return;

            const data = await res.json();
            renderCompositions(data);
        } catch {
            // ignore
        }
    }

    function renderCompositions(compositions) {
        if (!compositions || compositions.length === 0) {
            compositionsList.innerHTML = `
                <div class="empty-state">
                    <span class="empty-icon">🎵</span>
                    <p>No compositions yet. Go vent something!</p>
                </div>
            `;
            return;
        }

        compositionsList.innerHTML = compositions.map(comp => {
            const styleMap = {
                funny:   { emoji: '😂', name: 'Funny' },
                harsh:   { emoji: '🤘', name: 'Harsh' },
                hiphop:  { emoji: '🎤', name: 'Hip-Hop' },
                pansori: { emoji: '🥁', name: '판소리' },
            };
            const styleInfo = styleMap[comp.music_style] || { emoji: '🎵', name: comp.music_style };
            const styleEmoji = styleInfo.emoji;
            const styleName = styleInfo.name;
            const inputPreview = comp.input_type === 'clips'
                ? `🎙️ ${comp.input_text || 'Generated from clips'}`
                : comp.input_text
                    ? truncate(comp.input_text, 100)
                    : `🎤 Audio input`;
            const date = new Date(comp.created_at).toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
            });

            const actions = comp.status === 'done'
                ? `<div class="comp-actions">
                       <button class="btn-play" onclick="window.vibeApp.play('${comp.id}', '${styleName}')">▶ Play</button>
                       <button class="btn-download" onclick="window.vibeApp.download('${comp.id}')">⬇ Download</button>
                   </div>`
                : comp.status === 'failed'
                    ? `<span style="color:#ef4444;font-size:0.8rem">${comp.error_message ? truncate(comp.error_message, 50) : 'Error'}</span>`
                    : '';

            const lyricsSection = comp.generated_lyrics
                ? `<div class="comp-lyrics-toggle">
                       <button class="btn-lyrics" onclick="window.vibeApp.toggleLyrics('${comp.id}')">📝 Lyrics</button>
                   </div>
                   <div class="comp-lyrics" id="lyrics-${comp.id}" style="display:none">
                       <pre class="lyrics-text">${escapeHtml(comp.generated_lyrics)}</pre>
                   </div>`
                : '';

            return `
                <div class="composition-item" id="comp-${comp.id}">
                    <div class="comp-header">
                        <span class="comp-style">${styleEmoji} ${styleName}</span>
                        <span class="comp-voice">${comp.voice_gender && comp.voice_gender !== 'any' ? (comp.voice_gender === 'male' ? '🧔' : '👩') + ' ' + comp.voice_gender : ''}</span>
                        <span class="comp-status ${comp.status}">${comp.status}</span>
                    </div>
                    <div class="comp-input">${escapeHtml(inputPreview)}</div>
                    ${lyricsSection}
                    <div class="comp-meta">
                        <span class="comp-date">${date}</span>
                        ${actions}
                    </div>
                </div>
            `;
        }).join('');
    }

    // --- Player ---
    function playComposition(id, title) {
        const url = `/api/compositions/${id}/download`;
        playerAudio.src = '';

        // We need to set auth header via fetch for basic auth
        fetch(url, {
            headers: { 'Authorization': authHeader },
        })
            .then(res => {
                if (!res.ok) throw new Error('Failed to load audio');
                return res.blob();
            })
            .then(blob => {
                const objectUrl = URL.createObjectURL(blob);
                playerAudio.src = objectUrl;
                playerTitle.textContent = title || 'Now Playing';
                audioPlayer.classList.remove('hidden');
                playerAudio.play();
            })
            .catch(err => {
                alert('Error playing audio: ' + err.message);
            });
    }

    function downloadComposition(id) {
        fetch(`/api/compositions/${id}/download`, {
            headers: { 'Authorization': authHeader },
        })
            .then(res => {
                if (!res.ok) throw new Error('Download failed');
                return res.blob();
            })
            .then(blob => {
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = `vibe-${id.substring(0, 8)}.mp3`;
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
            })
            .catch(err => {
                alert('Download error: ' + err.message);
            });
    }

    function closePlayer() {
        playerAudio.pause();
        playerAudio.src = '';
        audioPlayer.classList.add('hidden');
    }

    function toggleLyrics(compositionId) {
        const el = document.getElementById(`lyrics-${compositionId}`);
        if (!el) return;
        const isHidden = el.style.display === 'none';
        el.style.display = isHidden ? 'block' : 'none';
        // Update button text
        const btn = el.previousElementSibling?.querySelector('.btn-lyrics');
        if (btn) {
            btn.textContent = isHidden ? '📝 Hide Lyrics' : '📝 Lyrics';
        }
    }

    // --- Notification ---
    function showNotification(msg) {
        // Simple notification via a temporary element
        const el = document.createElement('div');
        el.style.cssText = `
            position: fixed; top: 20px; right: 20px; z-index: 9999;
            padding: 12px 24px; border-radius: 12px;
            background: linear-gradient(135deg, #7c3aed, #ec4899);
            color: white; font-weight: 600; font-size: 0.9rem;
            box-shadow: 0 8px 32px rgba(124, 58, 237, 0.4);
            animation: fadeInUp 0.3s ease;
            font-family: 'Inter', sans-serif;
        `;
        el.textContent = msg;
        document.body.appendChild(el);
        setTimeout(() => {
            el.style.opacity = '0';
            el.style.transition = 'opacity 0.3s ease';
            setTimeout(() => el.remove(), 300);
        }, 3000);
    }

    // --- Helpers ---
    function truncate(str, max) {
        if (!str) return '';
        return str.length > max ? str.substring(0, max) + '...' : str;
    }

    function escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    function formatDuration(ms) {
        if (!ms || ms < 0) return '0:00';
        const totalSec = Math.floor(ms / 1000);
        const min = Math.floor(totalSec / 60);
        const sec = totalSec % 60;
        return `${min}:${sec.toString().padStart(2, '0')}`;
    }

    // --- Event Binding ---
    function bindEvents() {
        // Login
        loginForm.addEventListener('submit', handleLogin);
        logoutBtn.addEventListener('click', handleLogout);

        // Tabs
        tabBtns.forEach(btn => {
            btn.addEventListener('click', () => switchTab(btn.dataset.tab));
        });

        // Style selection
        document.querySelectorAll('input[name="style"]').forEach(radio => {
            radio.addEventListener('change', updateStyleSelection);
        });

        // Voice selection
        document.querySelectorAll('input[name="voice"]').forEach(radio => {
            radio.addEventListener('change', updateVoiceSelection);
        });

        // Lyric type selection
        document.querySelectorAll('input[name="lyric_type"]').forEach(radio => {
            radio.addEventListener('change', updateLyricTypeSelection);
        });

        // Record tab style selections
        ['rec_style', 'rec_voice', 'rec_lyric_type'].forEach(groupName => {
            document.querySelectorAll(`input[name="${groupName}"]`).forEach(radio => {
                radio.addEventListener('change', () => updateRecOptionSelection(groupName));
            });
        });

        // Generate
        generateBtn.addEventListener('click', handleCompose);

        // Record
        recordBtn.addEventListener('click', toggleRecording);
        generateFromClipsBtn.addEventListener('click', handleGenerateFromClips);

        // Library
        refreshBtn.addEventListener('click', loadCompositions);

        // Player
        playerCloseBtn.addEventListener('click', closePlayer);
    }

    // --- Expose for inline onclick handlers ---
    window.vibeApp = {
        play: playComposition,
        download: downloadComposition,
        toggleLyrics: toggleLyrics,
        playClip: playClip,
        deleteClip: deleteClip,
    };

    // --- Start ---
    init();
})();

