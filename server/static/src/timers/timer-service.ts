import type { ReactiveController, ReactiveControllerHost } from 'lit';
import type { ExpandedTimer } from '../types/boardgame-types.js';

export type TimerCadence = 'frame' | 'second';
export type TimerStatus = 'idle' | 'unavailable' | 'running' | 'elapsed';

export type TimerReference = ExpandedTimer;

export interface TimerReading {
  readonly timerId: string;
  readonly status: TimerStatus;
  readonly timeLeftMs: number;
  readonly durationMs: number;
  readonly secondsLeft: number;
  readonly progress: number;
}

export interface TimerControllerOptions {
  readonly cadence?: TimerCadence;
  /** Re-evaluated during host updates for controls whose presentation changes cadence. */
  readonly getCadence?: () => TimerCadence;
}

interface TimerInfoLike {
  readonly TimeLeft: number;
  readonly originalTimeLeft?: number;
}

interface TimerSubscription {
  readonly cadence: TimerCadence;
  readonly notify: (reading: TimerReading) => void;
  signature: string;
}

export const TIMER_SERVICE_REQUEST_EVENT = 'boardgame-timer-service-request';

export interface TimerServiceRequestDetail {
  readonly accept: (service: TimerService) => void;
}

const unavailableReading = (timerId = ''): TimerReading => Object.freeze({
  timerId,
  status: 'unavailable',
  timeLeftMs: 0,
  durationMs: 0,
  secondsLeft: 0,
  progress: 0,
});

const idleReading = (): TimerReading => Object.freeze({
  timerId: '',
  status: 'idle',
  timeLeftMs: 0,
  durationMs: 0,
  secondsLeft: 0,
  progress: 0,
});

/** Route-scoped adapter from the Redux clock to selective timer consumers. */
export class TimerService {
  private _readings = new Map<string, TimerReading>();
  private readonly _subscriptions = new Map<string, Set<TimerSubscription>>();

  update(infos: Readonly<Record<string, TimerInfoLike>> | null | undefined): void {
    const next = new Map<string, TimerReading>();
    for (const [timerId, info] of Object.entries(infos ?? {})) {
      if (!timerId) throw new Error('timer service: timer IDs must be non-empty strings');
      const timeLeftMs = info.TimeLeft;
      const durationMs = info.originalTimeLeft ?? timeLeftMs;
      if (!Number.isFinite(timeLeftMs) || timeLeftMs < 0) {
        throw new Error(`timer service: ${timerId} TimeLeft must be a finite non-negative number`);
      }
      if (!Number.isFinite(durationMs) || durationMs < 0 || timeLeftMs > durationMs) {
        throw new Error(`timer service: ${timerId} originalTimeLeft must be finite and at least TimeLeft`);
      }
      next.set(timerId, Object.freeze({
        timerId,
        status: timeLeftMs > 0 ? 'running' : 'elapsed',
        timeLeftMs,
        durationMs,
        secondsLeft: Math.ceil(timeLeftMs / 1000),
        progress: durationMs > 0 ? timeLeftMs / durationMs : 0,
      }));
    }
    this._readings = next;
    for (const [timerId, subscriptions] of this._subscriptions) {
      const reading = next.get(timerId) ?? unavailableReading(timerId);
      for (const subscription of subscriptions) this._notify(subscription, reading);
    }
  }

  subscribe(timerId: string, cadence: TimerCadence, notify: (reading: TimerReading) => void): () => void {
    if (!timerId.trim()) throw new Error('timer service: subscriptions require a non-empty timer ID');
    if (cadence !== 'frame' && cadence !== 'second') {
      throw new Error(`timer service: unknown cadence ${JSON.stringify(cadence)}`);
    }
    const subscription: TimerSubscription = { cadence, notify, signature: '' };
    let subscriptions = this._subscriptions.get(timerId);
    if (!subscriptions) {
      subscriptions = new Set();
      this._subscriptions.set(timerId, subscriptions);
    }
    subscriptions.add(subscription);
    this._notify(subscription, this._readings.get(timerId) ?? unavailableReading(timerId));
    return () => {
      subscriptions?.delete(subscription);
      if (subscriptions?.size === 0) this._subscriptions.delete(timerId);
    };
  }

  private _notify(subscription: TimerSubscription, reading: TimerReading): void {
    const signature = subscription.cadence === 'frame'
      ? `${reading.status}:${reading.timeLeftMs}:${reading.durationMs}`
      : `${reading.status}:${reading.secondsLeft}:${reading.durationMs}`;
    if (signature === subscription.signature) return;
    subscription.signature = signature;
    subscription.notify(reading);
  }
}

/** Lit controller for custom renderer UI that needs one live timer reading. */
export class TimerController implements ReactiveController {
  readonly host: ReactiveControllerHost & EventTarget;
  readonly getTimer: () => TimerReference | null | undefined;
  readonly cadence: TimerCadence;
  readonly getCadence: () => TimerCadence;
  private _reading: TimerReading = idleReading();

  private _service: TimerService | null = null;
  private _timerId = '';
  private _subscribedCadence: TimerCadence | null = null;
  private _unsubscribe: (() => void) | null = null;

  constructor(
    host: ReactiveControllerHost & EventTarget,
    getTimer: () => TimerReference | null | undefined,
    options: TimerControllerOptions = {},
  ) {
    this.host = host;
    this.getTimer = getTimer;
    if (options.cadence !== undefined && options.getCadence !== undefined) {
      throw new Error('timer controller: provide cadence or getCadence, not both');
    }
    this.cadence = options.cadence ?? 'second';
    this.getCadence = options.getCadence ?? (() => this.cadence);
    if (this.cadence !== 'frame' && this.cadence !== 'second') {
      throw new Error(`timer controller: unknown cadence ${JSON.stringify(this.cadence)}`);
    }
    host.addController(this);
  }

  get reading(): TimerReading {
    return this._reading;
  }

  hostConnected(): void {
    let accepted = false;
    this.host.dispatchEvent(new CustomEvent<TimerServiceRequestDetail>(TIMER_SERVICE_REQUEST_EVENT, {
      bubbles: true,
      composed: true,
      detail: {
        accept: service => {
          if (accepted) throw new Error('timer controller: more than one timer service accepted the request');
          accepted = true;
          this._service = service;
        },
      },
    }));
    if (!accepted || !this._service) {
      throw new Error('timer controller: no boardgame timer service was found; mount the renderer inside boardgame-game-view');
    }
    this._syncSubscription();
  }

  hostUpdate(): void {
    this._syncSubscription();
  }

  hostDisconnected(): void {
    this._unsubscribe?.();
    this._unsubscribe = null;
    this._timerId = '';
    this._subscribedCadence = null;
    this._service = null;
  }

  private _syncSubscription(): void {
    if (!this._service) return;
    const timer = this.getTimer();
    if (timer == null) {
      this._replaceSubscription('');
      return;
    }
    if (timer.IsTimer !== true || typeof timer.ID !== 'string') {
      throw new Error('timer controller: timer must be a generated timer reference with IsTimer=true and a string ID');
    }
    this._replaceSubscription(timer.ID, this.getCadence());
  }

  private _replaceSubscription(timerId: string, cadence: TimerCadence = this.getCadence()): void {
    if (cadence !== 'frame' && cadence !== 'second') {
      throw new Error(`timer controller: unknown cadence ${JSON.stringify(cadence)}`);
    }
    if (timerId === this._timerId && cadence === this._subscribedCadence) return;
    this._unsubscribe?.();
    this._unsubscribe = null;
    this._timerId = timerId;
    this._subscribedCadence = timerId ? cadence : null;
    if (!timerId || !this._service) {
      this._reading = idleReading();
      return;
    }
    let subscribing = true;
    this._unsubscribe = this._service.subscribe(timerId, cadence, reading => {
      this._reading = reading;
      if (!subscribing) this.host.requestUpdate();
    });
    subscribing = false;
  }
}
