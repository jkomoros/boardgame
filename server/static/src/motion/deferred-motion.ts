type FrameRequest = (callback: FrameRequestCallback) => number;
type FrameCancel = (handle: number) => void;

/**
 * Owns work deferred until two browser frames have elapsed. Scheduling new
 * work supersedes the old callback, and detached owners never start motion.
 */
export class DeferredMotion {
  private _generation = 0;
  private _frame: number | null = null;
  private readonly _requestFrame: FrameRequest;
  private readonly _cancelFrame: FrameCancel;

  constructor(
    requestFrame: FrameRequest = callback => requestAnimationFrame(callback),
    cancelFrame: FrameCancel = handle => cancelAnimationFrame(handle),
  ) {
    this._requestFrame = requestFrame;
    this._cancelFrame = cancelFrame;
  }

  schedule(owner: Element, callback: () => void): void {
    this.cancel();
    const generation = this._generation;
    const current = () => this._generation === generation && owner.isConnected;
    this._frame = this._requestFrame(() => {
      this._frame = null;
      if (!current()) return;
      this._frame = this._requestFrame(() => {
        this._frame = null;
        if (current()) callback();
      });
    });
  }

  cancel(): void {
    this._generation++;
    if (this._frame !== null) this._cancelFrame(this._frame);
    this._frame = null;
  }
}
