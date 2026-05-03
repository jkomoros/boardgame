/**
 * boardgame-chat-panel
 *
 * Collapsible chat drawer for multiplayer games. Renders a message list,
 * channel tabs (all/team), and a text input (or chip picker for pre-baked
 * messages).
 *
 * Auto-detects chat availability from the server's chat endpoint response.
 * If the server doesn't support chat, the panel doesn't render.
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state, query } from 'lit/decorators.js';
import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
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
    }

    .chat-header {
      display: flex;
      align-items: center;
      padding: 8px 12px;
      cursor: pointer;
      user-select: none;
      background: var(--md-sys-color-surface-container-high, #ece6f0);
    }

    .chat-header h4 {
      margin: 0;
      flex: 1;
      font-size: 14px;
      font-weight: 500;
    }

    .badge {
      background: var(--md-sys-color-error, #b3261e);
      color: var(--md-sys-color-on-error, #fff);
      border-radius: 10px;
      padding: 2px 6px;
      font-size: 11px;
      margin-left: 8px;
    }

    .chat-body {
      max-height: 300px;
      display: flex;
      flex-direction: column;
    }

    .chat-body.collapsed {
      max-height: 0;
      overflow: hidden;
    }

    .messages {
      flex: 1;
      overflow-y: auto;
      padding: 8px 12px;
      max-height: 220px;
      min-height: 60px;
    }

    .message {
      margin-bottom: 6px;
      font-size: 13px;
      line-height: 1.4;
    }

    .message .sender {
      font-weight: 500;
      color: var(--md-sys-color-primary, #6750a4);
    }

    .message.system {
      text-align: center;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      font-style: italic;
      font-size: 12px;
    }

    .input-area {
      display: flex;
      gap: 4px;
      padding: 8px 12px;
      border-top: 1px solid var(--md-sys-color-outline-variant, #cac4d0);
    }

    .input-area md-filled-text-field {
      flex: 1;
      --md-filled-text-field-container-shape: 20px;
    }

    .chip-area {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
      padding: 8px 12px;
      border-top: 1px solid var(--md-sys-color-outline-variant, #cac4d0);
    }

    .disabled-message {
      padding: 8px 12px;
      text-align: center;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      font-size: 12px;
      font-style: italic;
    }

    .empty-state {
      padding: 16px;
      text-align: center;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      font-size: 13px;
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
  private _collapsed = false;

  @state()
  private _unreadCount = 0;

  @state()
  private _lastMessageID = '';

  @query('.messages')
  private _messagesContainer!: HTMLElement;

  private _pollInterval: number | null = null;

  connectedCallback() {
    super.connectedCallback();
    this._fetchMessages();
    // Poll every 3 seconds for new messages (until WebSocket chat is wired)
    this._pollInterval = window.setInterval(() => this._fetchMessages(), 3000);

    // Listen for WebSocket chat notifications
    window.addEventListener('chat-notification', this._handleChatNotification as EventListener);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    if (this._pollInterval) {
      clearInterval(this._pollInterval);
      this._pollInterval = null;
    }
    window.removeEventListener('chat-notification', this._handleChatNotification as EventListener);
  }

  private _handleChatNotification = (e: CustomEvent) => {
    // A new chat message arrived via WebSocket — fetch it
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
        this._messages = [...this._messages, ...newMessages].slice(-100); // keep last 100
        this._lastMessageID = newMessages[newMessages.length - 1].id;

        if (this._collapsed) {
          this._unreadCount += newMessages.length;
        }

        // Auto-scroll to bottom
        this.updateComplete.then(() => {
          if (this._messagesContainer) {
            this._messagesContainer.scrollTop = this._messagesContainer.scrollHeight;
          }
        });
      }
    } catch {
      // Silently fail — chat is optional
    }
  }

  private _toggleCollapsed() {
    this._collapsed = !this._collapsed;
    if (!this._collapsed) {
      this._unreadCount = 0;
    }
  }

  private _senderName(senderIndex: number): string {
    if (senderIndex === -2) return 'System';
    if (senderIndex >= 0 && senderIndex < this.playersInfo.length) {
      return this.playersInfo[senderIndex]?.DisplayName || `Player ${senderIndex}`;
    }
    return `Player ${senderIndex}`;
  }

  private async _sendMessage(body: string) {
    if (!this.gameRoute || !body.trim()) return;

    try {
      const formData = new URLSearchParams();
      formData.append('channel', 'all');
      formData.append('body', body.trim());

      await fetch(`/api/game/${this.gameRoute.name}/${this.gameRoute.id}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData.toString(),
      });

      // Fetch immediately to show our own message
      await this._fetchMessages();
    } catch {
      // Silently fail
    }
  }

  private _handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      const input = e.target as HTMLInputElement & { value: string };
      const body = input.value;
      if (body.trim()) {
        this._sendMessage(body);
        input.value = '';
      }
    }
  }

  private _handleChipClick(msg: string) {
    this._sendMessage(msg);
  }

  render() {
    // Don't render if chat is not available
    if (this._chatConfig && !this._chatConfig.Enabled) return nothing;
    // Don't render until we've fetched config (first fetch may not have happened yet)
    if (!this.gameRoute) return nothing;

    const isObserver = this.viewingAsPlayer === -1;
    const isDisabled = isObserver || (this._chatConfig && !this._chatConfig.Enabled);
    const isPrebaked = this._chatConfig?.PrebakedOnly ?? false;
    const allowedMessages = this._chatConfig?.AllowedMessages ?? [];

    return html`
      <div class="chat-container">
        <div class="chat-header" @click=${this._toggleCollapsed}>
          <h4>Chat</h4>
          ${this._unreadCount > 0 ? html`<span class="badge">${this._unreadCount}</span>` : nothing}
          <md-icon>${this._collapsed ? 'expand_more' : 'expand_less'}</md-icon>
        </div>

        <div class="chat-body ${this._collapsed ? 'collapsed' : ''}">
          <div class="messages" role="log" aria-live="polite">
            ${this._messages.length === 0 ? html`
              <div class="empty-state">No messages yet</div>
            ` : this._messages.map(msg => html`
              <div class="message ${msg.sender === -2 ? 'system' : ''}">
                ${msg.sender === -2
                  ? html`${msg.body}`
                  : html`<span class="sender">${this._senderName(msg.sender)}:</span> ${msg.body}`
                }
              </div>
            `)}
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
                @keydown=${this._handleKeydown}>
              </md-filled-text-field>
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
