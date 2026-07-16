import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import firebase from 'firebase/compat/app';
import 'firebase/compat/auth';
import {
  PRIMARIES,
  randomAvatarIdentity,
  glyphForSlug,
} from './companion-avatar-catalog.js';
import { fauxSignInAsGuest } from '../actions/user.js';
import { apiHttpGet, apiHttpPost, buildApiUrl } from '../api.js';
import { gamePath } from '../util.js';
import {
  decodeJoinResponse,
  decodeJoinSeatResponse,
  decodeSeatOptionsResponse,
  type JoinResponse,
  type SeatOptionsResponse,
} from '../types/join-response.js';

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
type Step = 'code' | 'identity' | 'avatar' | 'avatarCustomize' | 'seat' | 'submitting' | 'error';

// Avatar primaries + name vocabulary live in companion-avatar-catalog.ts —
// imported above. Swap that module to upgrade the catalog without changing
// this component.

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
    .primary-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 8px;
      margin: 16px 0;
    }
    .primary-tile {
      padding: 12px;
      font-size: 48px;
      text-align: center;
      border: 2px solid #ddd;
      border-radius: 8px;
      cursor: pointer;
    }
    .primary-tile.selected {
      border-color: #1a73e8;
      background: #f0f7ff;
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

  // _unsubscribeIdTokenChanged is the cleanup handle returned by Firebase's
  // onIdTokenChanged subscription. Stashed here so disconnectedCallback can
  // call it; otherwise the listener leaks across navigations.
  private _unsubscribeIdTokenChanged: (() => void) | null = null;

  private _popstateHandler = (e: PopStateEvent) => {
    if (e.state && e.state.step) {
      // After a mid-flow reload, history entries for later steps can
      // survive while _joinResponse (in-memory only) is gone. Restoring
      // such a step would strand the user on a panel whose buttons
      // silently no-op — fall back to the code step instead.
      if (e.state.step !== 'code' && !this._joinResponse) {
        this._step = 'code';
        return;
      }
      this._step = e.state.step;
    }
  };

  private _setStep(step: Step) {
    this._step = step;
    history.pushState({ step }, '', window.location.pathname + window.location.search);
  }

  override connectedCallback() {
    super.connectedCallback();
    window.addEventListener('popstate', this._popstateHandler);
    history.replaceState({ step: 'code' }, '', window.location.pathname + window.location.search);
    // Roll a default avatar+name for the front door.
    this._reroll();
    // If the URL carries "?code=XXXX" (the QR on the Table view encodes
    // /join?code=<room code>), prefill AND auto-submit: scanning the QR is
    // an unambiguous statement of intent, so don't make the phone tap Join
    // on a code it didn't type. Malformed values just prefill; a bad-but-
    // well-formed code falls back to the code step with the server's error.
    const params = new URLSearchParams(window.location.search);
    const codeParam = params.get('code');
    if (codeParam) {
      this._codeInput = codeParam.toUpperCase();
      if (/^[A-Za-z]{4,5}$/.test(codeParam)) {
        this._submitCode();
      }
    } else {
      // No QR param: prefill (but don't auto-submit) the last code this
      // tab validated, so a mid-flow reload is one tap to recover.
      try {
        const last = sessionStorage.getItem('join-last-code');
        if (last) this._codeInput = last;
      } catch { /* private mode */ }
    }

    // Firebase anon (and Google) ID tokens are JWTs with a 1-hour TTL.
    // Subscribe to onIdTokenChanged so a long-lived join flow (or a
    // future re-auth-protected request) sees the freshest token.
    // Without this, a join older than an hour can't make authenticated
    // calls like /api/join/seat-options or seat-claim retries.
    const OFFLINE_DEV_MODE = Boolean(CONFIG && CONFIG.offline_dev_mode);
    if (!OFFLINE_DEV_MODE) {
      this._unsubscribeIdTokenChanged = firebase.auth().onIdTokenChanged(async (user) => {
        if (!user) {
          this._firebaseUID = '';
          this._firebaseToken = '';
          return;
        }
        this._firebaseUID = user.uid;
        try {
          this._firebaseToken = await user.getIdToken();
        } catch (err) {
          console.warn('Failed to refresh Firebase ID token:', err);
        }
      });
    }
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    window.removeEventListener('popstate', this._popstateHandler);
    if (this._unsubscribeIdTokenChanged) {
      this._unsubscribeIdTokenChanged();
      this._unsubscribeIdTokenChanged = null;
    }
  }

  private _reroll() {
    const identity = randomAvatarIdentity();
    this._displayName = identity.name;
    this._avatarSlug = identity.slug;
  }

  private async _submitCode() {
    this._error = '';
    const code = this._codeInput.trim().toUpperCase();
    if (!code) {
      this._error = 'Enter the 4-letter code shown on the projector';
      return;
    }
    try {
      const response = await apiHttpPost(buildApiUrl('join'), { code });
      if (response.status === 429) {
        this._error = 'Too many requests — slow down and try again in a moment';
        return;
      }
      if (!response.data) {
        this._error = response.error || response.friendlyError || 'Room not found';
        return;
      }
      this._joinResponse = decodeJoinResponse(response.data);
      // Remember the validated code for this tab so a mid-flow reload
      // (which loses _joinResponse and falls back to the code step)
      // doesn't make the player squint at the projector again.
      try { sessionStorage.setItem('join-last-code', code); } catch { /* private mode */ }
      this._setStep('identity');
    } catch (e) {
      this._error = 'Unable to join: ' + (e instanceof Error ? e.message : String(e));
    }
  }

  private async _continueAsGuest() {
    this._error = '';
    try {
      const OFFLINE_DEV_MODE = Boolean(CONFIG && CONFIG.offline_dev_mode);
      if (OFFLINE_DEV_MODE) {
        // In offline dev mode we don't have Firebase — synthesize a UID and
        // replace the faux persisted identity, mirroring how
        // signInAnonymously() replaces the Firebase user in production. If
        // we skipped this, the game page's auth bootstrap would re-validate
        // as the previously signed-in faux user and orphan the seat claim.
        this._firebaseUID = 'anon-' + Math.random().toString(36).slice(2, 14);
        this._firebaseToken = 'dev-mode-token';
        fauxSignInAsGuest(this._firebaseUID, 'Guest');
      } else {
        const cred = await firebase.auth().signInAnonymously();
        if (!cred.user) {
          this._error = 'Anonymous sign-in failed';
          return;
        }
        this._firebaseUID = cred.user.uid;
        this._firebaseToken = await cred.user.getIdToken();
      }
      this._setStep('avatar');
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
    this._displayName = this._displayName.trim();
    if (!this._displayName) {
      this._error = 'Please enter a display name';
      return;
    }
    if (this._joinResponse.requiresSeatPicker) {
      this._setStep('seat');
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
      const response = await apiHttpGet(buildApiUrl('join/seat-options', { gameID: this._joinResponse.gameID }), {
        headers: {
          'Authorization': 'Bearer ' + this._firebaseToken,
        },
      });
      if (!response.data) {
        this._error = response.error || response.friendlyError || 'Failed to load seats';
        this._step = 'error';
        return;
      }
      const seatOptions = decodeSeatOptionsResponse(response.data);
      if (seatOptions.gameID !== this._joinResponse.gameID || seatOptions.gameName !== this._joinResponse.gameName) {
        throw new Error('Seat options identified a different game');
      }
      if (seatOptions.requiresSeatPicker !== this._joinResponse.requiresSeatPicker) {
        throw new Error('Seat options contradicted the room seat-picker policy');
      }
      if (seatOptions.slots.length !== this._joinResponse.maxPlayers) {
        throw new Error('Seat options did not contain exactly one slot per player');
      }
      this._seatOptions = seatOptions;
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
      const body: Record<string, unknown> = {
        gameID: this._joinResponse.gameID,
        uid: this._firebaseUID,
        displayName: this._displayName,
        avatarSlug: this._avatarSlug,
        seatPick: this._selectedSeat !== null ? this._selectedSeat : -1,
      };
      const response = await apiHttpPost(buildApiUrl('join/seat'), body, {
        headers: {
          'Authorization': 'Bearer ' + this._firebaseToken,
        },
      });
      if (!response.data) {
        this._error = response.error || response.friendlyError || 'Failed to join';
        this._step = this._joinResponse.requiresSeatPicker ? 'seat' : 'avatar';
        return;
      }
      const seated = decodeJoinSeatResponse(response.data);
      if (seated.gameID !== this._joinResponse.gameID || seated.gameName !== this._joinResponse.gameName) {
        throw new Error('Seat result identified a different game');
      }
      if (seated.playerIndex >= this._joinResponse.maxPlayers) {
        throw new Error('Seat result player index was outside this game');
      }
      // Navigate to the game's Hand view. The surface=hand cookie was set
      // by the server in the response above; the loader at
      // boardgame-render-game.ts will pick the -hand.ts renderer on load.
      window.location.href = gamePath(seated.gameName, seated.gameID);
    } catch (e) {
      this._error = 'Unable to claim seat: ' + (e instanceof Error ? e.message : String(e));
      this._step = this._joinResponse.requiresSeatPicker ? 'seat' : 'avatar';
    }
  }

  private _pickSeat(playerIndex: number) {
    this._selectedSeat = playerIndex;
    this._submitSeat();
  }

  // "Room is full" isn't a dead end: the game is real and public state is
  // watchable — offer the Table view as a spectator path.
  private get _roomFull(): boolean {
    return /room is full/i.test(this._error);
  }

  private _watchInstead() {
    if (!this._joinResponse) return;
    window.location.href = gamePath(this._joinResponse.gameName, this._joinResponse.gameID) + '?display=table';
  }

  override render() {
    return html`
      <h2>Join a game</h2>
      ${this._error ? html`
        <div class="error">${this._error}</div>
        ${this._roomFull && this._joinResponse ? html`
          <button @click=${this._watchInstead}>Watch this game instead</button>
        ` : ''}
      ` : ''}

      <div class="step ${this._step === 'code' ? '' : 'hidden'}">
        <p>Enter the 4-letter code shown on the shared screen.</p>
        <input
          class="code-input"
          maxlength="5"
          inputmode="text"
          autocapitalize="characters"
          autocomplete="off"
          autocorrect="off"
          spellcheck="false"
          pattern="[A-Za-z]{4,5}"
          .value=${this._codeInput}
          @input=${(e: Event) => { this._codeInput = (e.target as HTMLInputElement).value.toUpperCase(); }}
          @keydown=${(e: KeyboardEvent) => { if (e.key === 'Enter') { e.preventDefault(); this._submitCode(); } }}
          placeholder="ABCD"
        />
        <button class="primary" @click=${this._submitCode}>Join</button>
      </div>

      <div class="step ${this._step === 'identity' ? '' : 'hidden'}">
        ${this._joinResponse ? html`<p>Joining <strong>${this._joinResponse.gameDisplayName}</strong></p>` : ''}
        <button class="primary" @click=${this._continueAsGuest}>Continue as guest</button>
        <button disabled title="Google sign-in for the join flow is coming soon">Sign in with Google (coming soon)</button>
      </div>

      <div class="step ${this._step === 'avatar' ? '' : 'hidden'}">
        <div class="avatar-front-door">
          <div class="glyph">${glyphForSlug(this._avatarSlug)}</div>
          <div class="name">${this._displayName}</div>
          <button class="primary" @click=${this._acceptAvatarAndProceed}>Looks good — join!</button>
          <br />
          <button @click=${this._reroll}>Try another</button>
          <br />
          <button @click=${() => { this._step = 'avatarCustomize'; }}>Customize</button>
        </div>
      </div>

      <div class="step ${this._step === 'avatarCustomize' ? '' : 'hidden'}">
        <p>Pick your avatar</p>
        <div class="primary-grid" role="radiogroup" aria-label="Avatar selection">
          ${PRIMARIES.map(p => html`
            <div
              class="primary-tile ${this._avatarSlug === p ? 'selected' : ''}"
              role="radio"
              tabindex="0"
              aria-checked=${this._avatarSlug === p ? 'true' : 'false'}
              aria-label="Avatar ${p}"
              @click=${() => { this._avatarSlug = p; }}
              @keydown=${(e: KeyboardEvent) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); this._avatarSlug = p; } }}>
              ${p}
            </div>
          `)}
        </div>
        <p style="margin-top:24px">Edit your name</p>
        <input
          class="code-input"
          maxlength="24"
          .value=${this._displayName}
          @input=${(e: Event) => { this._displayName = (e.target as HTMLInputElement).value; }}
          style="font-size:18px;letter-spacing:0;text-transform:none;text-align:left;"
        />
        <button class="primary" @click=${this._acceptAvatarAndProceed}>Looks good — join!</button>
        <button @click=${() => { this._step = 'avatar'; }}>Back</button>
      </div>

      <div class="step ${this._step === 'seat' ? '' : 'hidden'}">
        <p>Pick a seat</p>
        ${this._seatOptions ? html`
          <div class="slot-grid">
            ${this._seatOptions.slots.map(slot => html`
              <div class="slot ${slot.filled ? 'filled' : ''}"
                   @click=${() => { if (!slot.filled) this._pickSeat(slot.playerIndex); }}>
                <div>${slot.label}</div>
                ${slot.filled ? html`
                  <small>${slot.avatarSlug ? glyphForSlug(slot.avatarSlug) + ' ' : ''}${slot.displayName || 'Taken'}</small>
                ` : html`<small class="open-label">open</small>`}
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
