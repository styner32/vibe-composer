// ============================================
// VIBE COMPOSER — Frontend Application
// ============================================

(function () {
    'use strict';

    // --- State ---
    let authHeader = '';
    let currentUsername = '';
    let currentMode = 'text'; // 'text' | 'audio'
    let selectedAudioFile = null;
    let pollingInterval = null;
    let activeCompositionId = null;

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
    const libraryTab = $('#library-tab');

    // Compose
    const modeTextBtn = $('#mode-text-btn');
    const modeAudioBtn = $('#mode-audio-btn');
    const textInputSection = $('#text-input-section');
    const audioInputSection = $('#audio-input-section');
    const ventText = $('#vent-text');
    const audioDropZone = $('#audio-drop-zone');
    const audioFileInput = $('#audio-file-input');
    const audioPreview = $('#audio-preview');
    const audioFileName = $('#audio-file-name');
    const removeAudioBtn = $('#remove-audio-btn');
    const styleFunnyLabel = $('#style-funny-label');
    const styleHarshLabel = $('#style-harsh-label');
    const generateBtn = $('#generate-btn');
    const composeError = $('#compose-error');
    const activeGeneration = $('#active-generation');
    const statusMessage = $('#status-message');

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
        libraryTab.classList.toggle('active', tabName === 'library');

        if (tabName === 'library') {
            loadCompositions();
        }
    }

    // --- Input Mode ---
    function switchMode(mode) {
        currentMode = mode;
        modeTextBtn.classList.toggle('active', mode === 'text');
        modeAudioBtn.classList.toggle('active', mode === 'audio');
        textInputSection.classList.toggle('active', mode === 'text');
        audioInputSection.classList.toggle('active', mode === 'audio');
    }

    // --- Audio Upload ---
    function handleAudioFile(file) {
        if (!file) return;
        if (file.size > 25 * 1024 * 1024) {
            alert('File too large. Max 25MB.');
            return;
        }
        selectedAudioFile = file;
        audioFileName.textContent = file.name;
        audioDropZone.classList.add('hidden');
        audioPreview.classList.remove('hidden');
    }

    function removeAudio() {
        selectedAudioFile = null;
        audioFileInput.value = '';
        audioDropZone.classList.remove('hidden');
        audioPreview.classList.add('hidden');
    }

    // --- Style Selection ---
    function updateStyleSelection() {
        const selected = document.querySelector('input[name="style"]:checked').value;
        styleFunnyLabel.classList.toggle('selected', selected === 'funny');
        styleHarshLabel.classList.toggle('selected', selected === 'harsh');
    }

    // --- Compose ---
    async function handleCompose() {
        const style = document.querySelector('input[name="style"]:checked').value;
        const text = ventText.value.trim();

        if (currentMode === 'text' && !text) {
            showComposeError('Please tell us what\'s bothering you!');
            return;
        }
        if (currentMode === 'audio' && !selectedAudioFile) {
            showComposeError('Please upload an audio file!');
            return;
        }

        // Show loading state
        setGenerateLoading(true);
        composeError.classList.add('hidden');

        const formData = new FormData();
        formData.append('style', style);

        if (currentMode === 'text') {
            formData.append('text', text);
        } else {
            formData.append('audio', selectedAudioFile);
        }

        try {
            const res = await api('/api/compose', {
                method: 'POST',
                body: formData,
            });

            const data = await res.json();

            if (res.ok) {
                activeCompositionId = data.id;
                ventText.value = '';
                removeAudio();
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
            const styleEmoji = comp.music_style === 'funny' ? '😂' : '🤘';
            const styleName = comp.music_style === 'funny' ? 'Funny' : 'Harsh';
            const inputPreview = comp.input_text
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

            return `
                <div class="composition-item" id="comp-${comp.id}">
                    <div class="comp-header">
                        <span class="comp-style">${styleEmoji} ${styleName}</span>
                        <span class="comp-status ${comp.status}">${comp.status}</span>
                    </div>
                    <div class="comp-input">${escapeHtml(inputPreview)}</div>
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

    // --- Event Binding ---
    function bindEvents() {
        // Login
        loginForm.addEventListener('submit', handleLogin);
        logoutBtn.addEventListener('click', handleLogout);

        // Tabs
        tabBtns.forEach(btn => {
            btn.addEventListener('click', () => switchTab(btn.dataset.tab));
        });

        // Input modes
        modeTextBtn.addEventListener('click', () => switchMode('text'));
        modeAudioBtn.addEventListener('click', () => switchMode('audio'));

        // Audio upload
        audioDropZone.addEventListener('click', () => audioFileInput.click());
        audioFileInput.addEventListener('change', (e) => {
            if (e.target.files[0]) handleAudioFile(e.target.files[0]);
        });
        removeAudioBtn.addEventListener('click', removeAudio);

        // Drag and drop
        audioDropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            audioDropZone.classList.add('dragover');
        });
        audioDropZone.addEventListener('dragleave', () => {
            audioDropZone.classList.remove('dragover');
        });
        audioDropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            audioDropZone.classList.remove('dragover');
            if (e.dataTransfer.files[0]) handleAudioFile(e.dataTransfer.files[0]);
        });

        // Style selection
        document.querySelectorAll('input[name="style"]').forEach(radio => {
            radio.addEventListener('change', updateStyleSelection);
        });

        // Generate
        generateBtn.addEventListener('click', handleCompose);

        // Library
        refreshBtn.addEventListener('click', loadCompositions);

        // Player
        playerCloseBtn.addEventListener('click', closePlayer);
    }

    // --- Expose for inline onclick handlers ---
    window.vibeApp = {
        play: playComposition,
        download: downloadComposition,
    };

    // --- Start ---
    init();
})();
