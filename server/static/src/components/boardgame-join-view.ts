import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import firebase from 'firebase/compat/app';
import 'firebase/compat/auth';

/**
 * boardgame-join-view is the phone-side flow for joining a Table+Hand
 * companion-mode game. Multi-step inline state machine:
 *
 *   code  → identity → avatar → seat? → seated
 *
 * - code: user types the 4-letter room code shown on the Table view.
 *   POST /api/join validates + returns game metadata.
 * - identity: Continue-as-guest (Firebase signInAnonymously) or Sign-in-
 *   with-Google (existing OAuth flow).
 * - avatar: V1 stub renders one randomized name + emoji, with Looks-good
 *   + Try-another. P2.5 / P2.6 polish into the word-bloom-style 4-tuple
 *   composite avatar.
 * - seat: only shown if the /api/join response says requiresSeatPicker.
 *   Phone calls GET /api/join/seat-options for the slot grid.
 * - seated: POST /api/join/seat → on 200, server sets the surface=hand
 *   cookie scoped to gameID. Component navigates to /game/<name>/<id>.
 *
 * The component is deliberately Redux-free: the join flow is short, sees
 * no shared state, and benefits from keeping the multi-step transitions
 * inline. Reduxification can come later if the flow grows.
 */
type Step = 'code' | 'identity' | 'avatar' | 'seat' | 'submitting' | 'error';

interface JoinResponse {
  gameID: string;
  gameName: string;
  gameDisplayName: string;
  minPlayers: number;
  maxPlayers: number;
  currentPlayers: number;
  requiresSeatPicker: boolean;
}

interface SeatOptionsSlot {
  playerIndex: number;
  label: string;
  filled: boolean;
  avatarSlug?: string;
  displayName?: string;
}

interface SeatOptionsResponse {
  gameID: string;
  gameName: string;
  slots: SeatOptionsSlot[];
  requiresSeatPicker: boolean;
}

// V1 placeholder avatar set. P2.6 replaces these with proper SVG primaries.
const PLACEHOLDER_AVATARS = ['🦊', '🐻', '🦁', '🐯', '🐸', '🐙', '🦄', '🐳', '🦉', '🐧'];
const PLACEHOLDER_ADJECTIVES = ['Brave', 'Clever', 'Sunny', 'Wild', 'Bright', 'Mighty', 'Calm', 'Bold'];
const PLACEHOLDER_NOUNS = ['Fox', 'Bear', 'Lion', 'Tiger', 'Frog', 'Octopus', 'Unicorn', 'Whale', 'Owl', 'Penguin'];

function randomFromArray<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

function randomDisplayName(): string {
  return randomFromArray(PLACEHOLDER_ADJECTIVES) + randomFromArray(PLACEHOLDER_NOUNS);
}

function randomAvatarSlug(): string {
  // P2.6 will replace with composite primary-decoration-corner-tint slug.
  // For V1 the slug is just a single emoji codepoint.
  return randomFromArray(PLACEHOLDER_AVATARS);
}

@customElement('boardgame-join-view')
export class BoardgameJoinView extends LitElement {
  static styles = css`
    :host {
      display: block;
      max-width: 480px;
      margin: 0 auto;
      padding: 24px 16px;
      font-family: var(--md-sys-typescale-body-medium-font, system-ui), sans-serif;
    }
    h2 {
      font-size: 24px;
      margin: 0 0 16px 0;
    }
    .step {
      display: block;
    }
    .step.hidden {
      display: none;
    }
    .code-input {
      font-size: 32px;
      text-align: center;
      letter-spacing: 8px;
      text-transform: uppercase;
      width: 100%;
      padding: 16px;
      box-sizing: border-box;
      border: 2px solid #ccc;
      border-radius: 8px;
      margin: 8px 0 16px 0;
    }
    button {
      font-size: 18px;
      padding: 12px 24px;
      margin: 4px;
      cursor: pointer;
      border-radius: 8px;
      border: 1px solid #888;
      background: #f7f7f7;
    }
    button.primary {
      background: #1a73e8;
      color: white;
      border-color: #1a73e8;
      font-weight: 600;
    }
    .error {
      color: #c62828;
      margin: 12px 0;
    }
    .avatar-front-door {
      text-align: center;
      padding: 24px;
    }
    .avatar-front-door .glyph {
      font-size: 96px;
      margin-bottom: 8px;
    }
    .avatar-front-door .name {
      font-size: 24px;
      font-weight: 600;
      margin-bottom: 24px;
    }
    .slot-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 12px;
    }
    .slot {
      padding: 16px;
      border: 2px solid #ddd;
      border-radius: 8px;
      cursor: pointer;
      text-align: center;
    }
    .slot.filled {
      cursor: not-allowed;
      opacity: 0.5;
    }
    .slot:hover:not(.filled) {
      background: #f0f7ff;
      border-color: #1a73e8;
    }
  `;

  @property({ type: String })
  pageExtra = '';

  @state() private _step: Step = 'code';
  @state() private _error = '';
  @state() private _codeInput = '';
  @state() private _joinResponse: JoinResponse | null = null;
  @state() private _seatOptions: SeatOptionsResponse | null = null;
  @state() private _displayName = '';
  @state() private _avatarSlug = '';
  @state() private _firebaseUID = '';
  @state() private _firebaseToken = '';
  @state() private _selectedSeat: number | null = null;

  override connectedCallback() {
    super.connectedCallback();
    // Roll a default avatar+name for the front door.
    this._reroll();
    // If pageExtra starts with "?code=XXXX", prefill (handy for QR codes).
    const params = new URLSearchParams(window.location.search);
    const codeParam = params.get('code');
    if (codeParam) {
      this._codeInput = codeParam.toUpperCase();
    }
  }

  private _reroll() {
    this._displayName = randomDisplayName();
    this._avatarSlug = randomAvatarSlug();
  }

  private async _submitCode() {
    this._error = '';
    const code = this._codeInput.trim().toUpperCase();
    if (!code) {
      this._error = 'Enter the 4-letter code shown on the projector';
      return;
    }
    try {
      const apiHost = ((window as any).CONFIG && (window as any).CONFIG.dev_host) || '';
      const res = await fetch(apiHost + '/api/join', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code }),
        credentials: 'include',
      });
      if (res.status === 429) {
        this._error = 'Too many requests — slow down and try again in a moment';
        return;
      }
      if (!res.ok) {
        const body = await res.json().catch(() => ({ error: 'Room not found' }));
        this._error = body.error || 'Room not found';
        return;
      }
      this._joinResponse = await res.json();
      this._step = 'identity';
    } catch (e) {
      this._error = 'Network error: ' + (e instanceof Error ? e.message : String(e));
    }
  }

  private async _continueAsGuest() {
    this._error = '';
    try {
      const OFFLINE_DEV_MODE = (window as any).CONFIG && (window as any).CONFIG.offline_dev_mode;
      if (OFFLINE_DEV_MODE) {
        // In offline dev mode we don't have Firebase — synthesize a UID.
        this._firebaseUID = 'anon-' + Math.random().toString(36).slice(2, 14);
        this._firebaseToken = 'dev-mode-token';
      } else {
        const cred = await firebase.auth().signInAnonymously();
        if (!cred.user) {
          this._error = 'Anonymous sign-in failed';
          return;
        }
        this._firebaseUID = cred.user.uid;
        this._firebaseToken = await cred.user.getIdToken();
      }
      this._step = 'avatar';
    } catch (e) {
      this._error = 'Sign-in failed: ' + (e instanceof Error ? e.message : String(e));
    }
  }

  private async _continueWithGoogle() {
    // V1 stub — falls back to anonymous in offline dev mode; real Google
    // sign-in for prod is left as a P2.5 follow-up (depends on the
    // existing signInWithGoogle thunk in actions/user.ts). For now nudge
    // the user to anonymous to keep the autonomous flow moving.
    this._error = 'Google sign-in for the join flow is coming soon — use Continue as guest for now.';
  }

  private _acceptAvatarAndProceed() {
    if (!this._joinResponse) return;
    if (this._joinResponse.requiresSeatPicker) {
      this._step = 'seat';
      this._fetchSeatOptions();
    } else {
      // Symmetric: auto-assign on the server side.
      this._selectedSeat = null;
      this._submitSeat();
    }
  }

  private async _fetchSeatOptions() {
    if (!this._joinResponse) return;
    try {
      const apiHost = ((window as any).CONFIG && (window as any).CONFIG.dev_host) || '';
      const params = new URLSearchParams({
        gameID: this._joinResponse.gameID,
        uid: this._firebaseUID,
        token: this._firebaseToken,
      });
      const res = await fetch(apiHost + '/api/join/seat-options?' + params.toString(), {
        credentials: 'include',
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({ error: 'Failed to load seats' }));
        this._error = body.error || 'Failed to load seats';
        this._step = 'error';
        return;
      }
      this._seatOptions = await res.json();
    } catch (e) {
      this._error = 'Network error: ' + (e instanceof Error ? e.message : String(e));
      this._step = 'error';
    }
  }

  private async _submitSeat() {
    if (!this._joinResponse) return;
    this._step = 'submitting';
    this._error = '';
    try {
      const apiHost = ((window as any).CONFIG && (window as any).CONFIG.dev_host) || '';
      const body: Record<string, unknown> = {
        gameID: this._joinResponse.gameID,
        uid: this._firebaseUID,
        token: this._firebaseToken,
        displayName: this._displayName,
        avatarSlug: this._avatarSlug,
        seatPick: this._selectedSeat !== null ? this._selectedSeat : -1,
      };
      const res = await fetch(apiHost + '/api/join/seat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
        credentials: 'include',
      });
      if (!res.ok) {
        const errBody = await res.json().catch(() => ({ error: 'Failed to join' }));
        this._error = errBody.error || 'Failed to join';
        this._step = this._joinResponse.requiresSeatPicker ? 'seat' : 'avatar';
        return;
      }
      const seated = await res.json();
      // Navigate to the game's Hand view. The surface=hand cookie was set
      // by the server in the response above; the loader at
      // boardgame-render-game.ts will pick the -hand.ts renderer on load.
      window.location.href = '/' + 'game/' + seated.gameName + '/' + seated.gameID;
    } catch (e) {
      this._error = 'Network error: ' + (e instanceof Error ? e.message : String(e));
      this._step = this._joinResponse.requiresSeatPicker ? 'seat' : 'avatar';
    }
  }

  private _pickSeat(playerIndex: number) {
    this._selectedSeat = playerIndex;
    this._submitSeat();
  }

  override render() {
    return html`
      <h2>Join a game</h2>
      ${this._error ? html`<div class="error">${this._error}</div>` : ''}

      <div class="step ${this._step === 'code' ? '' : 'hidden'}">
        <p>Enter the 4-letter code shown on the shared screen.</p>
        <input
          class="code-input"
          maxlength="5"
          .value=${this._codeInput}
          @input=${(e: Event) => { this._codeInput = (e.target as HTMLInputElement).value.toUpperCase(); }}
          placeholder="ABCD"
        />
        <button class="primary" @click=${this._submitCode}>Join</button>
      </div>

      <div class="step ${this._step === 'identity' ? '' : 'hidden'}">
        ${this._joinResponse ? html`<p>Joining <strong>${this._joinResponse.gameDisplayName}</strong></p>` : ''}
        <button class="primary" @click=${this._continueAsGuest}>Continue as guest</button>
        <button @click=${this._continueWithGoogle}>Sign in with Google</button>
      </div>

      <div class="step ${this._step === 'avatar' ? '' : 'hidden'}">
        <div class="avatar-front-door">
          <div class="glyph">${this._avatarSlug}</div>
          <div class="name">${this._displayName}</div>
          <button class="primary" @click=${this._acceptAvatarAndProceed}>Looks good — join!</button>
          <br />
          <button @click=${this._reroll}>Try another</button>
        </div>
      </div>

      <div class="step ${this._step === 'seat' ? '' : 'hidden'}">
        <p>Pick a seat</p>
        ${this._seatOptions ? html`
          <div class="slot-grid">
            ${this._seatOptions.slots.map(slot => html`
              <div class="slot ${slot.filled ? 'filled' : ''}"
                   @click=${() => { if (!slot.filled) this._pickSeat(slot.playerIndex); }}>
                <div>${slot.label}</div>
                ${slot.filled ? html`<small>${slot.displayName || 'Taken'}</small>` : ''}
              </div>
            `)}
          </div>
        ` : html`<p>Loading…</p>`}
      </div>

      <div class="step ${this._step === 'submitting' ? '' : 'hidden'}">
        <p>Joining…</p>
      </div>

      <div class="step ${this._step === 'error' ? '' : 'hidden'}">
        <p>Something went wrong: ${this._error}</p>
        <button @click=${() => { this._step = 'code'; this._error = ''; }}>Start over</button>
      </div>
    `;
  }
}
