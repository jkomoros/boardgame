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
  decodeJoinProblem,
  decodeSeatOptionsResponse,
  type JoinResponse,
  type SeatOptionsResponse,
} from '../types/join-response.js';
import { forgetAllCompanionSurfaces, rememberSurfaceForGame } from '../utils/companion-surface.js';
import { codeFromJoinRoute, JoinSessionScope, type JoinOperation } from '../join/join-session.js';

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
      background: white;
      font: inherit;
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
  route = '';

  @property({ type: Boolean })
  selected = false;

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
  @state() private _problemCode = '';
  @state() private _existingIdentityLabel = '';

  private _sessionScope = new JoinSessionScope();
  private _attemptID = '';
  private _authGeneration = 0;

  // _unsubscribeIdTokenChanged is the cleanup handle returned by Firebase's
  // onIdTokenChanged subscription. Stashed here so disconnectedCallback can
  // call it; otherwise the listener leaks across navigations.
  private _unsubscribeIdTokenChanged: (() => void) | null = null;

  private _setStep(step: Step) {
    this._step = step;
  }

  override connectedCallback() {
    super.connectedCallback();
    // Firebase anon (and Google) ID tokens are JWTs with a 1-hour TTL.
    // Subscribe to onIdTokenChanged so a long-lived join flow (or a
    // future re-auth-protected request) sees the freshest token.
    // Without this, a join older than an hour can't make authenticated
    // calls like /api/join/seat-options or seat-claim retries.
    const OFFLINE_DEV_MODE = Boolean(CONFIG && CONFIG.offline_dev_mode);
    if (!OFFLINE_DEV_MODE) {
      this._unsubscribeIdTokenChanged = firebase.auth().onIdTokenChanged(async (user) => {
        const authGeneration = ++this._authGeneration;
        if (!user) {
          this._firebaseUID = '';
          this._firebaseToken = '';
          this._existingIdentityLabel = '';
          return;
        }
        this._firebaseUID = '';
        this._firebaseToken = '';
        this._existingIdentityLabel = '';
        try {
          const token = await user.getIdToken();
          if (authGeneration !== this._authGeneration || firebase.auth().currentUser?.uid !== user.uid) return;
          this._firebaseUID = user.uid;
          this._firebaseToken = token;
          this._existingIdentityLabel = user.displayName || user.email || (user.isAnonymous ? 'returning guest' : 'signed-in player');
        } catch (err) {
          console.warn('Failed to refresh Firebase ID token:', err);
        }
      });
    }
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    this._deactivate();
    if (this._unsubscribeIdTokenChanged) {
      this._unsubscribeIdTokenChanged();
      this._unsubscribeIdTokenChanged = null;
    }
  }

  protected override updated(changed: Map<PropertyKey, unknown>) {
    if (changed.has('selected') || changed.has('route')) {
      if (this.selected) this._activate(this.route);
      else this._deactivate();
    }
    if (changed.has('_step')) {
      const heading = this.renderRoot.querySelector<HTMLElement>('.step:not(.hidden) [data-step-heading]');
      heading?.focus();
    }
  }

  private _activate(route: string) {
    const key = route || '?';
    if (!this._sessionScope.activate(key)) return;
    this._step = 'code';
    this._error = '';
    this._problemCode = '';
    this._joinResponse = null;
    this._seatOptions = null;
    this._selectedSeat = null;
    this._attemptID = crypto.randomUUID();
    this._reroll();

    if (Boolean(CONFIG && CONFIG.offline_dev_mode)) {
      try {
        const uid = localStorage.getItem('faux-firebase-email') || '';
        if (uid) {
          this._firebaseUID = uid;
          this._firebaseToken = 'dev-mode-token';
          this._existingIdentityLabel = localStorage.getItem('faux-firebase-display-name') || uid;
        }
      } catch { /* private mode */ }
    }

    const codeParam = codeFromJoinRoute(route);
    if (codeParam) {
      this._codeInput = codeParam.toUpperCase();
      if (/^[A-Za-z]{4,5}$/.test(codeParam)) void this._submitCode();
    } else {
      try { this._codeInput = sessionStorage.getItem('join-last-code') || ''; } catch { this._codeInput = ''; }
    }
  }

  private _deactivate() {
    this._sessionScope.deactivate();
  }

  private _startOver() {
    this._sessionScope.activate((this.route || '?') + '#restart-' + crypto.randomUUID());
    this._step = 'code';
    this._error = '';
    this._problemCode = '';
    this._joinResponse = null;
    this._seatOptions = null;
    this._selectedSeat = null;
    this._attemptID = crypto.randomUUID();
  }

  private _beginOperation(): JoinOperation {
    return this._sessionScope.begin();
  }

  private _isCurrent(operation: JoinOperation): boolean {
    return this.selected && this._sessionScope.isCurrent(operation);
  }

  private _recoveryStep(code: string | undefined, fallback: Step): Step {
    switch (code) {
      case 'JOIN_TICKET_REQUIRED':
      case 'JOIN_EXPIRED':
        return 'code';
      case 'AUTH_REQUIRED':
      case 'AUTH_INVALID':
        return 'identity';
      case 'ROOM_LOCKED':
      case 'ROOM_FINISHED':
      case 'ROOM_FULL':
        return 'error';
      default:
        return fallback;
    }
  }

  private _reroll() {
    const identity = randomAvatarIdentity();
    this._displayName = identity.name;
    this._avatarSlug = identity.slug;
  }

  private _handleAvatarKey(event: KeyboardEvent, slug: string) {
    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight' && event.key !== 'ArrowUp' && event.key !== 'ArrowDown') return;
    event.preventDefault();
    const current = PRIMARIES.indexOf(slug);
    const delta = event.key === 'ArrowLeft' || event.key === 'ArrowUp' ? -1 : 1;
    const next = PRIMARIES[(current + delta + PRIMARIES.length) % PRIMARIES.length];
    if (!next) return;
    this._avatarSlug = next;
    void this.updateComplete.then(() => {
      this.renderRoot.querySelector<HTMLElement>(`.primary-tile[data-avatar="${next}"]`)?.focus();
    });
  }

  private async _submitCode() {
    if (this._step === 'submitting') return;
    this._error = '';
    this._problemCode = '';
    const code = this._codeInput.trim().toUpperCase();
    if (!code) {
      this._error = 'Enter the 4- or 5-letter code shown on the shared screen';
      return;
    }
    const operation = this._beginOperation();
    this._step = 'submitting';
    try {
      const response = await apiHttpPost(buildApiUrl('join'), { code }, { signal: operation.controller.signal });
      if (!this._isCurrent(operation) || response.aborted) return;
      if (response.status === 429) {
        this._error = 'Too many requests — slow down and try again in a moment';
        this._step = 'code';
        return;
      }
      if (!response.data) {
        this._error = response.error || response.friendlyError || 'Room not found';
        this._problemCode = response.code || '';
        this._step = 'code';
        return;
      }
      this._joinResponse = decodeJoinResponse(response.data);
      if (this._joinResponse.availableSeats === 0) {
        this._problemCode = 'ROOM_FULL';
        this._error = 'There are no open seats. Returning players can continue to recover their seat, or you can watch.';
      }
      // Remember the validated code for this tab so a mid-flow reload
      // (which loses _joinResponse and falls back to the code step)
      // doesn't make the player squint at the projector again.
      try { sessionStorage.setItem('join-last-code', code); } catch { /* private mode */ }
      this._setStep('identity');
    } catch (e) {
      if (!this._isCurrent(operation)) return;
      this._error = 'Unable to join: ' + (e instanceof Error ? e.message : String(e));
      this._step = 'code';
    }
  }

  private async _continueAsGuest() {
    if (this._step !== 'identity') return;
    this._error = '';
    const operation = this._beginOperation();
    this._step = 'submitting';
    try {
      const OFFLINE_DEV_MODE = Boolean(CONFIG && CONFIG.offline_dev_mode);
      let uid: string;
      let token: string;
      if (OFFLINE_DEV_MODE) {
        // In offline dev mode we don't have Firebase — synthesize a UID and
        // replace the faux persisted identity, mirroring how
        // signInAnonymously() replaces the Firebase user in production. If
        // we skipped this, the game page's auth bootstrap would re-validate
        // as the previously signed-in faux user and orphan the seat claim.
        uid = 'anon-' + Math.random().toString(36).slice(2, 14);
        token = 'dev-mode-token';
        fauxSignInAsGuest(uid, 'Guest');
      } else {
        const cred = await firebase.auth().signInAnonymously();
        if (!cred.user) {
          this._error = 'Anonymous sign-in failed';
          this._step = 'identity';
          return;
        }
        uid = cred.user.uid;
        token = await cred.user.getIdToken();
      }
      if (!this._isCurrent(operation)) return;
      forgetAllCompanionSurfaces();
      this._firebaseUID = uid;
      this._firebaseToken = token;
      this._existingIdentityLabel = 'returning guest';
      this._setStep('avatar');
    } catch (e) {
      if (!this._isCurrent(operation)) return;
      this._error = 'Sign-in failed: ' + (e instanceof Error ? e.message : String(e));
      this._step = 'identity';
    }
  }

  private _continueWithExistingIdentity() {
    if (this._step !== 'identity' || !this._firebaseUID || !this._firebaseToken) return;
    this._error = '';
    this._setStep('avatar');
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
    const operation = this._beginOperation();
    try {
      const response = await apiHttpGet(buildApiUrl('join/seat-options', { gameID: this._joinResponse.gameID }), {
        headers: {
          'Authorization': 'Bearer ' + this._firebaseToken,
          'X-Boardgame-Join-Ticket': this._joinResponse.joinTicket,
        },
        signal: operation.controller.signal,
      });
      if (!this._isCurrent(operation) || response.aborted) return;
      if (!response.data) {
        this._error = response.error || response.friendlyError || 'Failed to load seats';
        this._problemCode = response.code || '';
        this._step = this._recoveryStep(response.code, 'error');
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
      if (!this._isCurrent(operation)) return;
      this._error = 'Network error: ' + (e instanceof Error ? e.message : String(e));
      this._step = 'error';
    }
  }

  private async _submitSeat() {
    if (!this._joinResponse) return;
    if (this._step === 'submitting') return;
    this._step = 'submitting';
    this._error = '';
    this._problemCode = '';
    const operation = this._beginOperation();
    try {
      const body: Record<string, unknown> = {
        gameID: this._joinResponse.gameID,
        uid: this._firebaseUID,
        displayName: this._displayName,
        avatarSlug: this._avatarSlug,
        seatPick: this._selectedSeat !== null ? this._selectedSeat : -1,
        attemptID: this._attemptID,
      };
      const response = await apiHttpPost(buildApiUrl('join/seat'), body, {
        headers: {
          'Authorization': 'Bearer ' + this._firebaseToken,
          'X-Boardgame-Join-Ticket': this._joinResponse.joinTicket,
        },
        signal: operation.controller.signal,
      });
      if (!this._isCurrent(operation) || response.aborted) return;
      if (!response.data) {
        this._error = response.error || response.friendlyError || 'Failed to join';
        this._problemCode = response.code || '';
        if (response.failureBody && response.code === 'SEAT_TAKEN') {
          try {
            const problem = decodeJoinProblem(response.failureBody);
            if (problem.slots && this._seatOptions) {
              this._seatOptions = { ...this._seatOptions, slots: problem.slots };
            }
          } catch {
            void this._fetchSeatOptions();
          }
        }
        this._step = this._recoveryStep(response.code, this._joinResponse.requiresSeatPicker ? 'seat' : 'avatar');
        return;
      }
      const seated = decodeJoinSeatResponse(response.data);
      if (seated.gameID !== this._joinResponse.gameID || seated.gameName !== this._joinResponse.gameName) {
        throw new Error('Seat result identified a different game');
      }
      if (seated.playerIndex >= this._joinResponse.maxPlayers) {
        throw new Error('Seat result player index was outside this game');
      }
      rememberSurfaceForGame(seated.gameID, 'hand');
      try { sessionStorage.removeItem('join-last-code'); } catch { /* private mode */ }
      // Navigate to the game's Hand view. The surface=hand cookie was set
      // by the server in the response above; the loader at
      // boardgame-render-game.ts will pick the -hand.ts renderer on load.
      window.location.href = gamePath(seated.gameName, seated.gameID);
    } catch (e) {
      if (!this._isCurrent(operation)) return;
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
    return this._problemCode === 'ROOM_FULL';
  }

  private _watchInstead() {
    if (!this._joinResponse) return;
    window.location.href = gamePath(this._joinResponse.gameName, this._joinResponse.gameID) + '?display=table';
  }

  override render() {
    return html`
      <h2>Join a game</h2>
      ${this._error ? html`
        <div class="error" role="alert">${this._error}</div>
        ${this._roomFull && this._joinResponse ? html`
          <button @click=${this._watchInstead}>Watch this game instead</button>
        ` : ''}
      ` : ''}

      <div class="step ${this._step === 'code' ? '' : 'hidden'}">
        <h3 data-step-heading tabindex="-1">Enter room code</h3>
        <form @submit=${(e: SubmitEvent) => { e.preventDefault(); void this._submitCode(); }}>
        <label for="room-code">Enter the 4- or 5-letter code shown on the shared screen.</label>
        <input
          id="room-code"
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
          placeholder="ABCD"
        />
        <button type="submit" class="primary">Join</button>
        </form>
      </div>

      <div class="step ${this._step === 'identity' ? '' : 'hidden'}">
        <h3 data-step-heading tabindex="-1">Choose how to join</h3>
        ${this._joinResponse ? html`<p>Joining <strong>${this._joinResponse.gameDisplayName}</strong></p>` : ''}
        ${this._existingIdentityLabel && this._firebaseUID && this._firebaseToken ? html`
          <button class="primary" @click=${this._continueWithExistingIdentity}>Continue as ${this._existingIdentityLabel}</button>
          <br />
        ` : ''}
        <button class=${this._existingIdentityLabel ? '' : 'primary'} @click=${this._continueAsGuest}>Use a new guest identity</button>
        <button disabled title="Google sign-in for the join flow is coming soon">Sign in with Google (coming soon)</button>
      </div>

      <div class="step ${this._step === 'avatar' ? '' : 'hidden'}">
        <h3 data-step-heading tabindex="-1">Choose your table identity</h3>
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
        <h3 data-step-heading tabindex="-1">Customize your table identity</h3>
        <p id="avatar-legend">Pick your avatar</p>
        <div class="primary-grid" role="radiogroup" aria-label="Avatar selection">
          ${PRIMARIES.map(p => html`
            <button type="button"
              class="primary-tile ${this._avatarSlug === p ? 'selected' : ''}"
              role="radio"
              data-avatar=${p}
              tabindex=${this._avatarSlug === p ? '0' : '-1'}
              aria-checked=${this._avatarSlug === p ? 'true' : 'false'}
              aria-label="Avatar ${p}"
              @click=${() => { this._avatarSlug = p; }}
              @keydown=${(e: KeyboardEvent) => this._handleAvatarKey(e, p)}>
              ${p}
            </button>
          `)}
        </div>
        <label for="display-name" style="display:block;margin-top:24px">Edit your name</label>
        <input
          id="display-name"
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
        <h3 data-step-heading tabindex="-1">Pick a seat</h3>
        ${this._seatOptions ? html`
          <div class="slot-grid" aria-label="Available seats">
            ${this._seatOptions.slots.map(slot => html`
              <button type="button" class="slot ${!slot.available ? 'filled' : ''}"
                   ?disabled=${!slot.available}
                   @click=${() => { if (slot.available) this._pickSeat(slot.playerIndex); }}>
                <div>${slot.label}</div>
                ${!slot.available ? html`
                  <small>${slot.avatarSlug ? glyphForSlug(slot.avatarSlug) + ' ' : ''}${slot.displayName || (slot.status === 'agent' ? 'Computer' : slot.status === 'closed' ? 'Closed' : 'Taken')}</small>
                ` : html`<small class="open-label">open</small>`}
              </button>
            `)}
          </div>
        ` : html`<p role="status" aria-live="polite">Loading seats…</p>`}
      </div>

      <div class="step ${this._step === 'submitting' ? '' : 'hidden'}">
        <h3 data-step-heading tabindex="-1">Working</h3>
        <p role="status" aria-live="polite">Joining…</p>
      </div>

      <div class="step ${this._step === 'error' ? '' : 'hidden'}">
        <h3 data-step-heading tabindex="-1">Could not join</h3>
        <button @click=${this._startOver}>Start over</button>
      </div>
    `;
  }
}
