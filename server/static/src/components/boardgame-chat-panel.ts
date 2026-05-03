/**
 * boardgame-chat-panel
 *
 * Collapsible chat panel for multiplayer games. Renders a message list,
 * channel tabs (all/team), and a text input (or chip picker for pre-baked
 * messages).
 *
 * Auto-detects chat availability from the server's chat endpoint response.
 * If the server doesn't support chat, the panel doesn't render.
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state, query } from 'lit/decorators.js';
import '@material/web/textfield/filled-text-field.js';
import '@material/web/icon/icon.js';
import '@material/web/iconbutton/icon-button.js';
import '@material/web/chips/assist-chip.js';

interface ChatMessage {
  id: string;
  channel: string;
  sender: number;
  body: string;
  timestamp: number;
}

interface ChatConfig {
  Enabled: boolean;
  PrebakedOnly: boolean;
  AllowedMessages: string[] | null;
}

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
      background: var(--md-sys-color-surface-container, #f3edf7);
      border-radius: 12px;
      margin: 8px 0;
      overflow: hidden;
      box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
    }

    .chat-header {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 10px 14px;
      cursor: pointer;
      user-select: none;
      background: var(--md-sys-color-surface-container-high, #ece6f0);
      transition: background 0.15s;
    }

    .chat-header:hover {
      background: var(--md-sys-color-surface-container-highest, #e6e0e9);
    }

    .chat-header .icon {
      font-size: 18px;
      color: var(--md-sys-color-on-surface-variant, #49454f);
    }

    .chat-header h4 {
      margin: 0;
      flex: 1;
      font-size: 14px;
      font-weight: 500;
      font-family: var(--md-sys-typescale-title-small-font, 'Roboto', sans-serif);
      color: var(--md-sys-color-on-surface, #1c1b1f);
    }

    .badge {
      background: var(--md-sys-color-error, #b3261e);
      color: var(--md-sys-color-on-error, #fff);
      border-radius: 10px;
      padding: 2px 7px;
      font-size: 11px;
      font-weight: 500;
      min-width: 14px;
      text-align: center;
    }

    .toggle-icon {
      color: var(--md-sys-color-on-surface-variant, #49454f);
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
      font-family: var(--md-sys-typescale-body-small-font, 'Roboto', sans-serif);
    }

    .message .sender {
      font-weight: 600;
      margin-right: 4px;
    }

    .message .time {
      font-size: 10px;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      margin-left: 6px;
      opacity: 0.7;
    }

    .message.self .sender {
      color: var(--md-sys-color-primary, #6750a4);
    }

    .message.other .sender {
      color: var(--md-sys-color-tertiary, #7d5260);
    }

    .channel-tabs {
      display: flex;
      gap: 2px;
      padding: 4px 10px;
      border-bottom: 1px solid var(--md-sys-color-outline-variant, #cac4d0);
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
      border: 1px solid var(--md-sys-color-outline-variant, #cac4d0);
      color: var(--md-sys-color-on-surface-variant, #49454f);
      transition: background 0.15s, color 0.15s;
    }

    .channel-tab:hover {
      background: var(--md-sys-color-surface-container-highest, #e6e0e9);
    }

    .channel-tab.active {
      background: var(--md-sys-color-primary-container, #eaddff);
      color: var(--md-sys-color-on-primary-container, #21005d);
      border-color: var(--md-sys-color-primary, #6750a4);
    }

    .message.system {
      text-align: center;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      font-style: italic;
      font-size: 12px;
      padding: 6px 0;
      opacity: 0.8;
    }

    .input-area {
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 6px 10px;
      border-top: 1px solid var(--md-sys-color-outline-variant, #cac4d0);
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
      border-top: 1px solid var(--md-sys-color-outline-variant, #cac4d0);
    }

    .disabled-message {
      padding: 10px 14px;
      text-align: center;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      font-size: 12px;
      font-style: italic;
      border-top: 1px solid var(--md-sys-color-outline-variant, #cac4d0);
    }

    .empty-state {
      padding: 20px;
      text-align: center;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      font-size: 13px;
      opacity: 0.6;
    }
  `;

  @property({ type: Object })
  gameRoute: GameRoute | null = null;

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Array })
  playersInfo: PlayerInfo[] = [];

  @state()
  private _messages: ChatMessage[] = [];

  @state()
  private _chatConfig: ChatConfig | null = null;

  @state()
  private _viewChannels: string[] = [];

  @state()
  private _activeChannel = 'all';

  @state()
  private _collapsed = false;

  @state()
  private _unreadCount = 0;

  @state()
  private _lastMessageID = '';

  @state()
  private _sending = false;

  @query('.messages')
  private _messagesContainer!: HTMLElement;

  @query('md-filled-text-field')
  private _inputField!: HTMLElement & { value: string; focus: () => void };

  private _pollInterval: number | null = null;
  private _hasWebSocket = false;
  private _lastGameRouteId = '';

  protected willUpdate(changedProperties: Map<string, unknown>): void {
    // Reset chat state when switching games
    if (changedProperties.has('gameRoute') && this.gameRoute) {
      const newId = this.gameRoute.id;
      if (this._lastGameRouteId && this._lastGameRouteId !== newId) {
        this._messages = [];
        this._lastMessageID = '';
        this._unreadCount = 0;
        this._chatConfig = null;
        this._viewChannels = [];
        this._fetchMessages();
      }
      this._lastGameRouteId = newId;
    }
  }

  connectedCallback() {
    super.connectedCallback();
    this._fetchMessages();
    // Poll for messages — reduce frequency once WebSocket is working
    this._pollInterval = window.setInterval(() => {
      if (!this._hasWebSocket) this._fetchMessages();
    }, 3000);

    window.addEventListener('chat-notification', this._handleChatNotification as EventListener);
    window.addEventListener('socket-active', this._handleSocketActive as EventListener);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    if (this._pollInterval) {
      clearInterval(this._pollInterval);
      this._pollInterval = null;
    }
    window.removeEventListener('chat-notification', this._handleChatNotification as EventListener);
    window.removeEventListener('socket-active', this._handleSocketActive as EventListener);
  }

  private _handleSocketActive = () => {
    this._hasWebSocket = true; // Stop polling once WebSocket is confirmed working
  };

  private _handleChatNotification = () => {
    this._hasWebSocket = true;
    this._fetchMessages();
  };

  private async _fetchMessages() {
    if (!this.gameRoute) return;

    try {
      const url = `/api/game/${this.gameRoute.name}/${this.gameRoute.id}/chat?since=${this._lastMessageID}&limit=50`;
      const resp = await fetch(url);
      if (!resp.ok) return;

      const data = await resp.json();
      if (data.Status !== 'Success') return;

      this._chatConfig = data.ChatConfig || null;
      this._viewChannels = data.ViewChannels || [];

      const newMessages: ChatMessage[] = data.Messages || [];
      if (newMessages.length > 0) {
        this._messages = [...this._messages, ...newMessages].slice(-100);
        this._lastMessageID = newMessages[newMessages.length - 1].id;

        if (this._collapsed) {
          this._unreadCount += newMessages.length;
        }

        this.updateComplete.then(() => {
          if (this._messagesContainer) {
            this._messagesContainer.scrollTop = this._messagesContainer.scrollHeight;
          }
        });
      }
    } catch {
      // Chat is optional — silent failure
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
      // "dm/userA/userB" — show the OTHER user's player name
      const parts = channel.split('/');
      // Find which part is not the current user and resolve to a player name
      for (let i = 0; i < this.playersInfo.length; i++) {
        const name = this.playersInfo[i]?.DisplayName;
        if (name && (parts[1] === name || parts[2] === name)) {
          // Show the other party's name
          return name;
        }
      }
      // Fallback: show abbreviated IDs
      return 'DM';
    }
    return channel;
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

  private async _sendMessage(body: string) {
    if (!this.gameRoute || !body.trim() || this._sending) return;
    this._sending = true;

    try {
      const formData = new URLSearchParams();
      formData.append('channel', this._activeChannel);
      formData.append('body', body.trim());

      const resp = await fetch(`/api/game/${this.gameRoute.name}/${this.gameRoute.id}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData.toString(),
      });

      if (resp.ok) {
        await this._fetchMessages();
      }
    } catch {
      // Silent failure
    } finally {
      this._sending = false;
    }
  }

  private _handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      this._submitInput();
    }
  }

  private _submitInput() {
    if (!this._inputField) return;
    const body = this._inputField.value;
    if (body.trim()) {
      this._sendMessage(body);
      this._inputField.value = '';
    }
  }

  private _handleSendClick() {
    this._submitInput();
  }

  private _handleChipClick(msg: string) {
    this._sendMessage(msg);
  }

  render() {
    if (this._chatConfig && !this._chatConfig.Enabled) return nothing;
    if (!this.gameRoute) return nothing;

    const isObserver = this.viewingAsPlayer === -1;
    const isDisabled = isObserver || (this._chatConfig && !this._chatConfig.Enabled);
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
          <md-icon class="toggle-icon ${this._collapsed ? 'collapsed' : ''}">expand_less</md-icon>
        </div>

        <div class="chat-body ${this._collapsed ? 'collapsed' : ''}">
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
            ` : this._filteredMessages.map(msg => {
              const isSelf = msg.sender === this.viewingAsPlayer;
              const isSystem = msg.sender === -2;
              return html`
                <div class="message ${isSystem ? 'system' : isSelf ? 'self' : 'other'}">
                  ${isSystem
                    ? html`— ${msg.body} —`
                    : html`<span class="sender">${this._senderName(msg.sender)}</span>${msg.body}<span class="time">${this._formatTime(msg.timestamp)}</span>`
                  }
                </div>
              `;
            })}
          </div>

          ${isDisabled ? html`
            <div class="disabled-message">
              ${isObserver ? 'Observers cannot send messages' : 'Chat is not available right now'}
            </div>
          ` : isPrebaked ? html`
            <div class="chip-area">
              ${allowedMessages.map(msg => html`
                <md-assist-chip
                  label=${msg}
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
