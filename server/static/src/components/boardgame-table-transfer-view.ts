import { LitElement, css, html } from 'lit';
import { customElement, property, query } from 'lit/decorators.js';
import { apiHttpPost, buildApiUrl } from '../api.js';
import { rememberSurfaceForGame, tableRecoveryDeviceID } from '../utils/companion-surface.js';
import {
  clearPendingTableTransfer,
  decodeTableTransferInspection, decodeTableTransferRedemption,
  rememberPendingTableTransfer, restorePendingTableTransfer,
  TableTransferScope, transferFailureMessage, transferTokenFromFragment,
  type TableTransferInput, type TableTransferInspection,
} from '../table-transfer/table-transfer.js';

@customElement('boardgame-table-transfer-view')
export class BoardgameTableTransferView extends LitElement {
  static override styles = css`
    :host { display: block; min-height: 100%; background: #f4f7fa; color: #17212b; }
    main { box-sizing: border-box; max-width: 620px; margin: 0 auto; padding: clamp(24px, 6vw, 64px) 20px; }
    .card { padding: clamp(20px, 5vw, 36px); border-radius: 16px; background: white; box-shadow: 0 8px 28px rgb(23 33 43 / 14%); }
    h1 { margin-top: 0; font-size: clamp(28px, 6vw, 44px); }
    form { display: grid; gap: 16px; }
    label { display: grid; gap: 6px; font-weight: 650; }
    input { box-sizing: border-box; width: 100%; min-height: 48px; padding: 10px 12px; border: 2px solid #718090; border-radius: 8px; font: inherit; text-transform: uppercase; }
    button { min-height: 48px; padding: 10px 18px; border: 0; border-radius: 8px; background: #245f94; color: white; font: inherit; font-weight: 700; cursor: pointer; }
    button.secondary { background: #e4ebf1; color: #17212b; }
    button:disabled { cursor: wait; opacity: .65; }
    .actions { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 20px; }
    .error { color: #a11616; font-weight: 650; }
    .hint { color: #52606d; }
  `;

  @property({ type: Boolean }) selected = false;
  @property({ type: Object, attribute: false }) private _inspection: TableTransferInspection | null = null;
  @property({ type: Boolean, attribute: false }) private _pending = false;
  @property({ type: String, attribute: false }) private _error = '';
  @property({ type: String, attribute: false }) private _roomCode = '';
  @property({ type: String, attribute: false }) private _manualCode = '';
  @property({ type: Number, attribute: false }) private _secondsRemaining = 0;
  @query('h1') private _heading?: HTMLElement;

  private readonly _scope = new TableTransferScope();
  private _input: TableTransferInput | null = null;
  private _fragmentConsumed = false;
  private _restoredConfirmed = false;
  private _expiryDeadline = 0;
  private _expiryTimer: ReturnType<typeof setInterval> | null = null;

  override connectedCallback(): void {
    super.connectedCallback();
    window.addEventListener('hashchange', this._handleHashChange);
    this._consumeFragment();
  }

  private readonly _handleHashChange = (): void => {
    if (!this.selected) return;
    this._scope.invalidate();
    this._pending = false;
    this._inspection = null;
    this._input = null;
    this._error = '';
    this._restoredConfirmed = false;
    this._stopExpiryTimer();
    this._fragmentConsumed = false;
    this._consumeFragment();
  };

  private _consumeFragment(): void {
    if (!this._fragmentConsumed) {
      this._fragmentConsumed = true;
      const token = transferTokenFromFragment(window.location.hash);
      if (window.location.hash) {
        window.history.replaceState(window.history.state, '', window.location.pathname + window.location.search);
      }
      if (token) {
        this._input = { kind: 'token', token };
        this._restoredConfirmed = false;
        rememberPendingTableTransfer(this._input, false);
        void this._inspect(this._input);
      } else {
        const restored = restorePendingTableTransfer();
        if (restored) {
          this._input = restored.input;
          this._restoredConfirmed = restored.confirmed;
          void this._inspect(restored.input);
        }
      }
    }
  }

  override disconnectedCallback(): void {
    window.removeEventListener('hashchange', this._handleHashChange);
    this._scope.invalidate();
    this._stopExpiryTimer();
    super.disconnectedCallback();
  }

  protected override updated(changed: Map<PropertyKey, unknown>): void {
    if (changed.has('selected')) {
      if (this.selected) {
        this._consumeFragment();
      } else {
        this._scope.invalidate();
        this._pending = false;
        this._inspection = null;
        this._input = null;
        this._error = '';
        this._restoredConfirmed = false;
        this._stopExpiryTimer();
        this._fragmentConsumed = false;
      }
    }
    if ((changed.has('_inspection') || changed.has('selected')) && this.selected) this._heading?.focus();
  }

  private _requestBody(input: TableTransferInput): Readonly<Record<string, unknown>> {
    return input.kind === 'token'
      ? { token: input.token }
      : { roomCode: input.roomCode, manualCode: input.manualCode };
  }

  private async _inspect(input: TableTransferInput): Promise<void> {
    const operation = this._scope.begin();
    this._pending = true;
    this._error = '';
    rememberPendingTableTransfer(input, this._restoredConfirmed);
    let resumeConfirmedRedemption = false;
    try {
      const response = await apiHttpPost(buildApiUrl('table-transfer/inspect'), this._requestBody(input), { signal: operation.signal });
      if (!operation.isCurrent()) return;
      if (!response.data) {
        this._error = transferFailureMessage(response.code, response.error || response.friendlyError);
        if (this._terminalFailure(response.code)) clearPendingTableTransfer();
        return;
      }
      const inspection = decodeTableTransferInspection(response.data);
      if (inspection.expiresAtMs <= inspection.serverNowMs) {
        this._error = transferFailureMessage('TABLE_TRANSFER_EXPIRED');
        clearPendingTableTransfer();
        return;
      }
      this._input = input;
      this._inspection = inspection;
      this._installExpiryTimer(inspection);
      resumeConfirmedRedemption = this._restoredConfirmed;
    } catch (error) {
      if (operation.isCurrent()) {
        console.error('[table-transfer] malformed inspect response', error);
        this._error = 'The server returned an invalid transfer response. Please try again.';
      }
    } finally {
      if (operation.isCurrent()) {
        this._pending = false;
        if (resumeConfirmedRedemption) void this._redeem();
      }
    }
  }

  private _submitManual(event: Event): void {
    event.preventDefault();
    const roomCode = this._roomCode.trim().toUpperCase();
    const manualCode = this._manualCode.trim().toUpperCase();
    if (!roomCode || !manualCode) {
      this._error = 'Enter both codes shown on the current shared Table.';
      return;
    }
    this._restoredConfirmed = false;
    void this._inspect({ kind: 'manual', roomCode, manualCode });
  }

  private async _redeem(): Promise<void> {
    const input = this._input;
    const inspection = this._inspection;
    if (!input || !inspection || this._pending) return;
    const operation = this._scope.begin();
    this._restoredConfirmed = true;
    rememberPendingTableTransfer(input, true);
    this._pending = true;
    this._error = '';
    try {
      const response = await apiHttpPost(buildApiUrl('table-transfer/redeem'), {
        ...this._requestBody(input), pairingID: inspection.pairingID,
        deviceID: tableRecoveryDeviceID(inspection.gameID),
      }, { signal: operation.signal });
      if (!operation.isCurrent()) return;
      if (!response.data) {
        this._error = transferFailureMessage(response.code, response.error || response.friendlyError);
        if (this._terminalFailure(response.code)) clearPendingTableTransfer();
        return;
      }
      const result = decodeTableTransferRedemption(response.data);
      if (result.gameID !== inspection.gameID || result.gameName !== inspection.gameName) {
        throw new Error('Redeemed game does not match the inspected transfer');
      }
      rememberSurfaceForGame(result.gameID, 'table');
      const target = new URL(result.gameURL, window.location.origin);
      target.searchParams.set('display', 'table');
      window.location.replace(target.pathname + target.search);
    } catch (error) {
      if (operation.isCurrent()) {
        console.error('[table-transfer] malformed redeem response', error);
        this._error = 'The server returned an invalid redemption response. Please try again.';
      }
    } finally {
      if (operation.isCurrent()) this._pending = false;
    }
  }

  private _startOver(): void {
    this._scope.invalidate();
    this._input = null;
    this._inspection = null;
    this._error = '';
    this._restoredConfirmed = false;
    this._stopExpiryTimer();
    clearPendingTableTransfer();
  }

  private _terminalFailure(code: string | undefined): boolean {
    return code === 'TABLE_TRANSFER_EXPIRED' || code === 'TABLE_TRANSFER_CANCELLED'
      || code === 'TABLE_TRANSFER_INVALID' || code === 'GAME_FINISHED';
  }

  private _installExpiryTimer(inspection: TableTransferInspection): void {
    this._stopExpiryTimer();
    this._expiryDeadline = Date.now() + Math.max(0, inspection.expiresAtMs - inspection.serverNowMs);
    this._updateExpiry();
    if (this._secondsRemaining > 0) this._expiryTimer = setInterval(() => this._updateExpiry(), 1000);
  }

  private _updateExpiry(): void {
    this._secondsRemaining = Math.max(0, Math.ceil((this._expiryDeadline - Date.now()) / 1000));
    if (this._secondsRemaining === 0 && this._inspection) {
      this._error = transferFailureMessage('TABLE_TRANSFER_EXPIRED');
      this._inspection = null;
      this._restoredConfirmed = false;
      clearPendingTableTransfer();
      this._stopExpiryTimer();
    }
  }

  private _stopExpiryTimer(): void {
    if (this._expiryTimer !== null) clearInterval(this._expiryTimer);
    this._expiryTimer = null;
  }

  private _formatTime(seconds: number): string {
    return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`;
  }

  override render() {
    return html`<main><section class="card" aria-busy=${this._pending ? 'true' : 'false'}>
      ${this._inspection ? html`
        <h1 tabindex="-1">${this._inspection.alreadyRedeemed ? `Reconnect ${this._inspection.gameDisplayName} here?` : `Move ${this._inspection.gameDisplayName} here?`}</h1>
        <p>${this._inspection.alreadyRedeemed
          ? 'This link already connected a screen. Continue only to restore that exact same screen after a lost response or reload.'
          : 'This screen will become the shared Table. The old screen will stop controlling the game.'}</p>
        <p aria-live="polite">Transfer expires in ${this._formatTime(this._secondsRemaining)}.</p>
        <div class="actions">
          <button type="button" ?disabled=${this._pending} @click=${this._redeem}>${this._pending ? 'Connecting…' : (this._inspection.alreadyRedeemed ? 'Reconnect this shared Table' : 'Make this the shared Table')}</button>
          <button type="button" class="secondary" ?disabled=${this._pending} @click=${this._startOver}>Use a different code</button>
        </div>
      ` : this._input?.kind === 'token' ? html`
        <h1 tabindex="-1">Connect a shared Table</h1>
        <p>${this._pending ? 'Checking the transfer link…' : 'The transfer link could not be checked.'}</p>
        ${!this._pending ? html`<div class="actions">
          <button type="button" @click=${() => { if (this._input) void this._inspect(this._input); }}>Try link again</button>
          <button type="button" class="secondary" @click=${this._startOver}>Enter codes instead</button>
        </div>` : ''}
      ` : html`
        <h1 tabindex="-1">Connect a shared Table</h1>
        <p>Enter the two codes shown after choosing “Move shared Table” on the current screen.</p>
        <form @submit=${this._submitManual}>
          <label>Room code<input autocomplete="off" autocapitalize="characters" inputmode="text" spellcheck="false" maxlength="8" .value=${this._roomCode} @input=${(e: Event) => { this._roomCode = (e.target as HTMLInputElement).value; }}></label>
          <label>Transfer code<input autocomplete="one-time-code" autocapitalize="characters" inputmode="text" spellcheck="false" maxlength="16" .value=${this._manualCode} @input=${(e: Event) => { this._manualCode = (e.target as HTMLInputElement).value; }}></label>
          <button type="submit" ?disabled=${this._pending}>${this._pending ? 'Checking…' : 'Continue'}</button>
        </form>
      `}
      ${this._error ? html`<p class="error" role="alert">${this._error}</p>` : html`<p class="hint" aria-live="polite">${this._pending ? 'Contacting the shared Table…' : ''}</p>`}
    </section></main>`;
  }
}

declare global { interface HTMLElementTagNameMap { 'boardgame-table-transfer-view': BoardgameTableTransferView; } }
