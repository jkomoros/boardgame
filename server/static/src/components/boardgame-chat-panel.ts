/**
 * boardgame-chat-panel
 *
 * Collapsible chat panel for multiplayer games. Renders a message list,
 * channel tabs (all/team), and a text input (or chip picker for pre-baked
 * messages).
 *
 * Auto-detects chat policy from the server's chat endpoint response. A
 * deliberately disabled policy hides the panel; transport and contract errors
 * stay visible with a retry instead of masquerading as an empty conversation.
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state, query } from 'lit/decorators.js';
import '@material/web/textfield/filled-text-field.js';
import '@material/web/icon/icon.js';
import '@material/web/iconbutton/icon-button.js';
import '@material/web/chips/assist-chip.js';
import { apiGet, apiPost, buildGameUrl } from '../api.js';
import {
  decodeChatReadResponse,
  decodeChatSendResponse,
  type ChatConfig,
  type ChatMessage,
} from '../types/chat-response.js';

interface PlayerInfo {
  IsEmpty: boolean;
  IsAgent: boolean;
  DisplayName: string;
}

interface GameRoute {
  name: string;
  id: string;
}

@customElement('boardgame-chat-panel')
export class BoardgameChatPanel extends LitElement {
  static styles = css`
    :host {
      display: block;
    }

    .chat-container {
      background: var(--md-sys-color-surface-container, #F0EBE3);
      border-radius: 12px;
      margin: 8px 0;
      overflow: hidden;
      box-shadow: 0 1px 3px rgba(60, 40, 20, 0.10),
                  inset 0 1px 0 rgba(255, 255, 255, 0.3);
    }

    .chat-header {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 10px 14px;
      cursor: pointer;
      user-select: none;
      background: linear-gradient(180deg, var(--md-sys-color-surface-container-high, #E8E2D8) 0%, var(--md-sys-color-surface-container, #F0EBE3) 100%);
      transition: background 0.2s ease;
    }

    .chat-header:hover {
      background: var(--md-sys-color-surface-container-highest, #E0D9CE);
    }

    .chat-header .icon {
      font-size: 18px;
      color: var(--md-sys-color-on-surface-variant, #4A4539);
    }

    .chat-header h4 {
      margin: 0;
      flex: 1;
      font-size: 14px;
      font-weight: 500;
      font-family: var(--md-sys-typescale-title-small-font, 'Source Sans 3', sans-serif);
      color: var(--md-sys-color-on-surface, #1C1810);
    }

    .badge {
      background: var(--md-sys-color-error, #BA1A1A);
      color: var(--md-sys-color-on-error, #fff);
      border-radius: 10px;
      padding: 2px 7px;
      font-size: 11px;
      font-weight: 500;
      min-width: 14px;
      text-align: center;
    }

    .toggle-icon {
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      transition: transform 0.2s;
    }

    .toggle-icon.collapsed {
      transform: rotate(180deg);
    }

    .chat-body {
      display: flex;
      flex-direction: column;
      max-height: 320px;
      transition: max-height 0.25s ease-in-out, opacity 0.2s;
      opacity: 1;
    }

    .chat-body.collapsed {
      max-height: 0;
      opacity: 0;
      overflow: hidden;
    }

    @media (prefers-reduced-motion: reduce) {
      .chat-body, .toggle-icon {
        transition: none;
      }
    }

    .messages {
      flex: 1;
      overflow-y: auto;
      padding: 10px 14px;
      max-height: 240px;
      min-height: 50px;
      scroll-behavior: smooth;
    }

    .message {
      margin-bottom: 4px;
      padding: 4px 0;
      font-size: 13px;
      line-height: 1.5;
      font-family: var(--md-sys-typescale-body-small-font, 'Source Sans 3', sans-serif);
    }

    .message .sender {
      font-weight: 600;
      margin-right: 4px;
    }

    .message .time {
      font-size: 10px;
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      margin-left: 6px;
      opacity: 0.7;
    }

    .message.self .sender {
      color: var(--md-sys-color-primary, #2E6B4F);
    }

    .message.other .sender {
      color: var(--md-sys-color-tertiary, #8B7432);
    }

    .channel-tabs {
      display: flex;
      gap: 2px;
      padding: 4px 10px;
      border-bottom: 1px solid var(--md-sys-color-outline-variant, #CCC4B8);
      overflow-x: auto;
      flex-shrink: 0;
    }

    .channel-tab {
      padding: 4px 10px;
      border-radius: 16px;
      font-size: 12px;
      cursor: pointer;
      white-space: nowrap;
      background: transparent;
      border: 1px solid var(--md-sys-color-outline-variant, #CCC4B8);
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      transition: background 0.2s ease, color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
    }

    .channel-tab:hover {
      background: var(--md-sys-color-surface-container-highest, #E0D9CE);
    }

    .channel-tab.active {
      background: var(--md-sys-color-primary-container, #D4E8DA);
      color: var(--md-sys-color-on-primary-container, #0A2818);
      border-color: var(--md-sys-color-primary, #2E6B4F);
      box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.06);
    }

    .message.system {
      text-align: center;
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      font-style: italic;
      font-size: 12px;
      padding: 6px 0;
      opacity: 0.8;
    }

    .message.grouped {
      margin-bottom: 1px;
      padding-top: 0;
    }

    .message.grouped .sender,
    .message.grouped .time {
      display: none;
    }

    .channel-indicator {
      font-size: 10px;
      padding: 1px 5px;
      border-radius: 6px;
      margin-left: 4px;
      background: var(--md-sys-color-surface-container-highest, #E0D9CE);
      color: var(--md-sys-color-on-surface-variant, #4A4539);
    }

    .channel-indicator.dm {
      background: var(--md-sys-color-tertiary-container, #F0E4C4);
      color: var(--md-sys-color-on-tertiary-container, #2E2508);
    }

    .channel-indicator.team {
      background: var(--md-sys-color-secondary-container, #EDDCC8);
      color: var(--md-sys-color-on-secondary-container, #271A10);
    }

    .input-area {
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 6px 10px;
      border-top: 1px solid var(--md-sys-color-outline-variant, #CCC4B8);
    }

    .input-area md-filled-text-field {
      flex: 1;
      --md-filled-text-field-container-shape: 20px;
      --md-filled-text-field-top-space: 8px;
      --md-filled-text-field-bottom-space: 8px;
    }

    .chip-area {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
      padding: 8px 14px;
      border-top: 1px solid var(--md-sys-color-outline-variant, #CCC4B8);
    }

    .disabled-message {
      padding: 10px 14px;
      text-align: center;
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      font-size: 12px;
      font-style: italic;
      border-top: 1px solid var(--md-sys-color-outline-variant, #CCC4B8);
    }

    .chat-error {
      padding: 8px 14px;
      color: var(--md-sys-color-error, #BA1A1A);
      background: var(--md-sys-color-error-container, #FFDAD6);
      font-size: 12px;
    }

    .chat-error button {
      margin-left: 8px;
    }

    .empty-state {
      padding: 20px;
      text-align: center;
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      font-size: 13px;
      opacity: 0.6;
    }
  `;

  @property({ type: Object, attribute: 'game-route' })
  gameRoute: GameRoute | null = null;

  @property({ type: Number, attribute: 'viewing-as-player' })
  viewingAsPlayer = 0;

  @property({ type: Array, attribute: 'players-info' })
  playersInfo: PlayerInfo[] = [];

  @state()
  private _messages: ChatMessage[] = [];

  @state()
  private _chatConfig: ChatConfig | null = null;

  @state()
  private _viewChannels: string[] = [];

  @state()
  private _sendChannels: string[] = [];

  @state()
  private _activeChannel = 'all';

  @state()
  private _userIDMap: Record<string, number> = {};

  @state()
  private _collapsed = false;

  @state()
  private _unreadCount = 0;

  @state()
  private _lastMessageID = '';

  @state()
  private _sending = false;

  @state()
  private _chatError = '';

  @state()
  private _sendError = '';

  @state()
  private _soundEnabled = false;

  @query('.messages')
  private _messagesContainer!: HTMLElement;

  @query('md-filled-text-field')
  private _inputField!: HTMLElement & { value: string; focus: () => void };

  private _pollInterval: number | null = null;
  private _hasWebSocket = false;
  private _lastGameRouteKey = '';
  private _fetchController: AbortController | null = null;
  private _fetchPromise: Promise<void> | null = null;

  protected willUpdate(changedProperties: Map<string, unknown>): void {
    // Fetch messages on first gameRoute set, and reset state when switching games
    if (changedProperties.has('gameRoute') && !this.gameRoute) {
      this._fetchController?.abort();
      this._lastGameRouteKey = '';
      this._hasWebSocket = false;
      this._messages = [];
      this._lastMessageID = '';
      this._unreadCount = 0;
      this._chatConfig = null;
      this._viewChannels = [];
      this._sendChannels = [];
      this._activeChannel = 'all';
      this._userIDMap = {};
      this._chatError = '';
      this._sendError = '';
    } else if (changedProperties.has('gameRoute') && this.gameRoute) {
      const newKey = this._routeKey(this.gameRoute);
      if (this._lastGameRouteKey !== newKey) {
        this._fetchController?.abort();
        if (this._lastGameRouteKey) {
          // Switching games — reset state
          this._messages = [];
          this._lastMessageID = '';
          this._unreadCount = 0;
          this._chatConfig = null;
          this._viewChannels = [];
          this._sendChannels = [];
          this._activeChannel = 'all';
          this._userIDMap = {};
          this._chatError = '';
          this._sendError = '';
        }
        this._lastGameRouteKey = newKey;
        this._hasWebSocket = false;
        void this._fetchMessages(true);
      }
    }
  }

  connectedCallback() {
    super.connectedCallback();
    // Load sound preference
    try {
      this._soundEnabled = localStorage.getItem('boardgame-chat-sound') === '1';
    } catch { /* ignore */ }

    void this._fetchMessages();
    // Poll for messages — reduce frequency once WebSocket is working
    this._pollInterval = window.setInterval(() => {
      if (!this._hasWebSocket) void this._fetchMessages();
    }, 3000);

    window.addEventListener('chat-notification', this._handleChatNotification as EventListener);
    window.addEventListener('socket-active', this._handleSocketActive as EventListener);
    window.addEventListener('socket-active-changed', this._handleSocketActiveChanged as EventListener);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    if (this._pollInterval) {
      clearInterval(this._pollInterval);
      this._pollInterval = null;
    }
    this._fetchController?.abort();
    window.removeEventListener('chat-notification', this._handleChatNotification as EventListener);
    window.removeEventListener('socket-active', this._handleSocketActive as EventListener);
    window.removeEventListener('socket-active-changed', this._handleSocketActiveChanged as EventListener);
  }

  private _handleSocketActive = () => {
    this._hasWebSocket = true; // Stop polling once WebSocket is confirmed working
  };

  private _handleSocketActiveChanged = (event: Event) => {
    if (!(event instanceof CustomEvent)) return;
    const detail: unknown = event.detail;
    if (detail === null || typeof detail !== 'object' || Array.isArray(detail)) return;
    const value = Reflect.get(detail, 'value');
    if (typeof value === 'boolean') this._hasWebSocket = value;
  };

  private _handleChatNotification = () => {
    this._hasWebSocket = true;
    void this._fetchMessages(true);
  };

  private _routeKey(route: GameRoute): string {
    return `${route.name}\u0000${route.id}`;
  }

  private async _fetchMessages(force = false): Promise<void> {
    if (!this.gameRoute) return;
    if (this._fetchPromise) {
      if (!force) return this._fetchPromise;
      this._fetchController?.abort();
      await this._fetchPromise;
    }
    const route = this.gameRoute;
    const routeKey = this._routeKey(route);
    const controller = new AbortController();
    this._fetchController = controller;
    const request = this._runMessageFetch(route, routeKey, controller);
    this._fetchPromise = request;
    try {
      await request;
    } finally {
      if (this._fetchPromise === request) this._fetchPromise = null;
      if (this._fetchController === controller) this._fetchController = null;
    }
  }

  private async _runMessageFetch(route: GameRoute, routeKey: string, controller: AbortController): Promise<void> {
    const response = await apiGet<unknown>(buildGameUrl(route.name, route.id, 'chat', {
      since: this._lastMessageID,
      limit: 50,
    }), controller.signal);
    if (controller.signal.aborted || !this.gameRoute || this._routeKey(this.gameRoute) !== routeKey) return;
    if (!response.data) {
      this._chatError = response.error || response.friendlyError || 'Chat could not be refreshed';
      return;
    }
    try {
      const data = decodeChatReadResponse(response.data);
      this._chatConfig = data.ChatConfig;
      this._viewChannels = data.ViewChannels;
      this._sendChannels = data.SendChannels;
      this._userIDMap = data.UserIDMap;
      this._chatError = '';
      if (!data.ViewChannels.includes(this._activeChannel)) {
        this._activeChannel = data.ViewChannels.includes('all') ? 'all' : (data.ViewChannels[0] ?? 'all');
      }

      if (data.Messages.length === 0) return;
      this._lastMessageID = data.Messages[data.Messages.length - 1].id;
      const known = new Set(this._messages.map(message => message.id));
      const newMessages = data.Messages.filter(message => !known.has(message.id));
      if (newMessages.length === 0) return;
      this._messages = [...this._messages, ...newMessages].slice(-100);

      if (this._collapsed) this._unreadCount += newMessages.length;
      if (newMessages.some(message => message.sender !== this.viewingAsPlayer && message.sender !== -2)) {
        this._playNotificationSound();
      }
      void this.updateComplete.then(() => {
        if (this._messagesContainer) this._messagesContainer.scrollTop = this._messagesContainer.scrollHeight;
      });
    } catch (error) {
      console.error('[chat] rejected server payload:', error);
      this._chatError = error instanceof Error ? error.message : 'Chat returned an invalid response';
    }
  }

  private _toggleCollapsed() {
    this._collapsed = !this._collapsed;
    if (!this._collapsed) {
      this._unreadCount = 0;
      // Focus the input when expanding
      this.updateComplete.then(() => this._inputField?.focus());
    }
  }

  /** Get a human-readable display name for a channel */
  private _channelDisplayName(channel: string): string {
    if (channel === 'all') return 'All';
    if (channel.startsWith('team/')) return channel.substring(5);
    if (channel.startsWith('dm/')) {
      // "dm/userIdA/userIdB" — find the OTHER user's player index via UserIDMap
      const parts = channel.split('/');
      const otherUserID = parts[1] === this._myUserID() ? parts[2] : parts[1];
      const playerIdx = this._userIDMap[otherUserID];
      if (playerIdx !== undefined && playerIdx >= 0 && playerIdx < this.playersInfo.length) {
        return this.playersInfo[playerIdx]?.DisplayName || `Player ${playerIdx}`;
      }
      return 'DM';
    }
    return channel;
  }

  /** Get the current user's ID from the UserIDMap (the one matching viewingAsPlayer) */
  private _myUserID(): string {
    for (const [uid, idx] of Object.entries(this._userIDMap)) {
      if (idx === this.viewingAsPlayer) return uid;
    }
    return '';
  }

  /** Get messages filtered to the active channel (or all if "all") */
  private get _filteredMessages(): ChatMessage[] {
    if (this._activeChannel === 'all') {
      // Show all messages the user can see
      return this._messages;
    }
    return this._messages.filter(m => m.channel === this._activeChannel);
  }

  /** Whether we have multiple channels to show tabs for */
  private get _hasMultipleChannels(): boolean {
    return this._viewChannels.length > 1;
  }

  private _selectChannel(channel: string) {
    this._activeChannel = channel;
    this._sendError = '';
  }

  private _senderName(senderIndex: number): string {
    if (senderIndex === -2) return 'System';
    if (senderIndex >= 0 && senderIndex < this.playersInfo.length) {
      return this.playersInfo[senderIndex]?.DisplayName || `Player ${senderIndex}`;
    }
    return `Player ${senderIndex}`;
  }

  private _formatTime(timestamp: number): string {
    if (!timestamp) return '';
    const d = new Date(timestamp);
    const h = d.getHours();
    const m = d.getMinutes().toString().padStart(2, '0');
    const ampm = h >= 12 ? 'PM' : 'AM';
    const h12 = h % 12 || 12;
    return `${h12}:${m} ${ampm}`;
  }

  private async _sendMessage(body: string): Promise<boolean> {
    if (!this.gameRoute || !body.trim() || this._sending) return false;
    if (!this._sendChannels.includes(this._activeChannel)) {
      this._sendError = 'You cannot send messages to this channel';
      return false;
    }
    const route = this.gameRoute;
    const routeKey = this._routeKey(route);
    this._sending = true;
    this._sendError = '';

    try {
      const response = await apiPost<unknown>(
        buildGameUrl(route.name, route.id, 'chat'),
        { channel: this._activeChannel, body: body.trim() },
        'application/x-www-form-urlencoded',
      );
      if (!this.gameRoute || this._routeKey(this.gameRoute) !== routeKey) return false;
      if (!response.data) {
        this._sendError = response.friendlyError || response.error || 'Message could not be sent';
        return false;
      }
      decodeChatSendResponse(response.data);
      await this._fetchMessages(true);
      return true;
    } catch (error) {
      console.error('[chat] send failed:', error);
      this._sendError = error instanceof Error ? error.message : 'Message could not be sent';
      return false;
    } finally {
      this._sending = false;
    }
  }

  private _handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      void this._submitInput();
    }
  }

  private async _submitInput(): Promise<void> {
    if (!this._inputField) return;
    const body = this._inputField.value;
    if (body.trim()) {
      if (await this._sendMessage(body)) this._inputField.value = '';
    }
  }

  private _handleSendClick() {
    void this._submitInput();
  }

  private _handleChipClick(msg: string) {
    void this._sendMessage(msg);
  }

  private _toggleSound() {
    this._soundEnabled = !this._soundEnabled;
    try {
      localStorage.setItem('boardgame-chat-sound', this._soundEnabled ? '1' : '0');
    } catch { /* localStorage not available */ }
  }

  /** Play a short notification sound via Web Audio API */
  private _playNotificationSound() {
    if (!this._soundEnabled) return;
    try {
      const ctx = new AudioContext();
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.frequency.setValueAtTime(800, ctx.currentTime);
      osc.frequency.setValueAtTime(600, ctx.currentTime + 0.08);
      gain.gain.setValueAtTime(0.15, ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.15);
      osc.start(ctx.currentTime);
      osc.stop(ctx.currentTime + 0.15);
      // Close AudioContext after sound finishes to prevent resource leak
      osc.onended = () => ctx.close();
    } catch { /* Audio not available */ }
  }

  /** Check if a message is grouped with the previous one (same sender within 60s) */
  private _isGrouped(messages: ChatMessage[], index: number): boolean {
    if (index === 0) return false;
    const prev = messages[index - 1];
    const curr = messages[index];
    if (curr.sender === -2 || prev.sender === -2) return false; // system messages break groups
    if (curr.sender !== prev.sender) return false;
    if (curr.channel !== prev.channel) return false;
    // Group if within 60 seconds
    return Math.abs(curr.timestamp - prev.timestamp) < 60000;
  }

  /** Get the channel indicator badge for a message when viewing "all" */
  private _channelBadge(channel: string): string {
    if (channel === 'all') return '';
    if (channel.startsWith('team/')) return channel.substring(5);
    if (channel.startsWith('dm/')) return 'DM';
    return channel;
  }

  render() {
    if (this._chatConfig && !this._chatConfig.Enabled) return nothing;
    if (!this.gameRoute) return nothing;
    if (!this._chatConfig && !this._chatError) return nothing;

    const isObserver = this.viewingAsPlayer === -1;
    const canSendToActiveChannel = this._sendChannels.includes(this._activeChannel);
    const isDisabled = isObserver || !canSendToActiveChannel;
    const isPrebaked = this._chatConfig?.PrebakedOnly ?? false;
    const allowedMessages = this._chatConfig?.AllowedMessages ?? [];

    return html`
      <div class="chat-container">
        <div class="chat-header" @click=${this._toggleCollapsed}
             role="button" tabindex="0" aria-expanded=${!this._collapsed}
             @keydown=${(e: KeyboardEvent) => e.key === 'Enter' && this._toggleCollapsed()}>
          <md-icon class="icon">chat</md-icon>
          <h4>Chat</h4>
          ${this._unreadCount > 0 ? html`
            <span class="badge" aria-label="${this._unreadCount} unread messages">${this._unreadCount}</span>
          ` : nothing}
          <md-icon-button @click=${(e: Event) => { e.stopPropagation(); this._toggleSound(); }}
                          aria-label="${this._soundEnabled ? 'Mute notifications' : 'Enable sound notifications'}"
                          title="${this._soundEnabled ? 'Sound on' : 'Sound off'}">
            <md-icon>${this._soundEnabled ? 'volume_up' : 'volume_off'}</md-icon>
          </md-icon-button>
          <md-icon class="toggle-icon ${this._collapsed ? 'collapsed' : ''}">expand_less</md-icon>
        </div>

        <div class="chat-body ${this._collapsed ? 'collapsed' : ''}">
          ${this._chatError ? html`
            <div class="chat-error" role="status">
              Chat unavailable: ${this._chatError}
              <button @click=${() => { void this._fetchMessages(true); }}>Retry</button>
            </div>
          ` : nothing}
          ${this._hasMultipleChannels ? html`
            <div class="channel-tabs" role="tablist" aria-label="Chat channels">
              ${this._viewChannels.map(ch => html`
                <button class="channel-tab ${ch === this._activeChannel ? 'active' : ''}"
                        role="tab" aria-selected=${ch === this._activeChannel ? 'true' : 'false'}
                        @click=${() => this._selectChannel(ch)}>
                  ${this._channelDisplayName(ch)}
                </button>
              `)}
            </div>
          ` : nothing}

          <div class="messages" role="log" aria-live="polite" aria-label="Chat messages">
            ${this._filteredMessages.length === 0 ? html`
              <div class="empty-state">No messages yet. Say hello!</div>
            ` : this._filteredMessages.map((msg, idx) => {
              const isSelf = msg.sender === this.viewingAsPlayer;
              const isSystem = msg.sender === -2;
              const grouped = this._isGrouped(this._filteredMessages, idx);
              const showChannelBadge = this._activeChannel === 'all' && this._hasMultipleChannels;
              const badge = showChannelBadge ? this._channelBadge(msg.channel) : '';
              return html`
                <div class="message ${isSystem ? 'system' : isSelf ? 'self' : 'other'} ${grouped ? 'grouped' : ''}">
                  ${isSystem
                    ? html`— ${msg.body} —`
                    : html`<span class="sender">${this._senderName(msg.sender)}</span>${msg.body}${badge ? html`<span class="channel-indicator ${msg.channel.startsWith('dm/') ? 'dm' : msg.channel.startsWith('team/') ? 'team' : ''}">${badge}</span>` : nothing}<span class="time">${this._formatTime(msg.timestamp)}</span>`
                  }
                </div>
              `;
            })}
          </div>

          ${isDisabled ? html`
            <div class="disabled-message">
              ${isObserver
                ? 'Observers cannot send messages'
                : 'This channel is read-only for you'}
            </div>
          ` : isPrebaked ? html`
            <div class="chip-area">
              ${allowedMessages.map(msg => html`
                <md-assist-chip
                  label=${msg}
                  ?disabled=${this._sending}
                  @click=${() => this._handleChipClick(msg)}>
                </md-assist-chip>
              `)}
            </div>
          ` : html`
            <div class="input-area">
              <md-filled-text-field
                placeholder="Type a message..."
                @keydown=${this._handleKeydown}
                ?disabled=${this._sending}
                aria-label="Chat message">
              </md-filled-text-field>
              <md-icon-button @click=${this._handleSendClick}
                              ?disabled=${this._sending}
                              aria-label="Send message">
                <md-icon>send</md-icon>
              </md-icon-button>
            </div>
          `}
          ${this._sendError ? html`<div class="chat-error" role="alert">${this._sendError}</div>` : nothing}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-chat-panel': BoardgameChatPanel;
  }
}
