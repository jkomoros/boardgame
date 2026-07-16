# Route-Scoped Resilient Live Sessions

## Problem

Game reads, Redux state, socket reconnects, and retry scheduling currently have
related but independent lifetimes. Navigating while an `/info` or `/version`
request is pending can leave the new route with the old route's data or loading
flags. Network failure can immediately retrigger reads and reconnect sockets in
an unbounded loop. The renderer exposes this as an indefinite blocking spinner.

## Invariants

- Changing `{name,id}` atomically clears every game-owned payload, animation,
  timer, request, error, version, and socket field.
- Only deliberate viewer preferences (`requestedPlayer` and
  `autoCurrentPlayer`) survive a game-route change.
- Every info/version request has an identity. Only the currently active request
  for the current route may settle Redux state.
- Starting a newer read aborts the older read of that kind. Route changes,
  deactivation, and disconnection abort both kinds.
- Retry work is bounded exponential backoff with jitter and a cap. Success,
  route change, and explicit retry reset the attempt count.
- Scheduled callbacks capture route identity and become inert when superseded.
- The rendered game remains visible during recoverable outages, but stale move
  actions are blocked and an accessible status explains connecting, retrying,
  or offline state.
- Retryable renderer-module failures offer an explicit retry; contract and
  registration failures remain loud and do not loop.

## Implementation slices

1. Atomic reducer route transition plus request IDs and abortable thunks.
2. Shared pure retry-delay policy; version and socket scheduling use it.
3. Explicit connection status passed through game view to renderer host, with
   online/offline integration and manual retry.
4. Explicit failed preview state and retryable renderer loading.

## Verification stories

- Deferred game A info/version reads, navigate to B, and prove B starts while A
  can neither settle flags nor install state/static information.
- Change admin/player perspective on the same route; the newest request wins.
- Fake-clock failures prove retry delay progression, cap, cancellation, and
  success reset without duplicate in-flight work.
- Socket callbacks and timers from route A cannot reconnect or mutate route B.
- Keyboard and screen-reader users receive useful non-modal recovery status and
  can retry; finished games suppress only their expected disconnect.
- A transient dynamic import failure succeeds on explicit retry, while missing
  registration remains a developer-visible permanent error.
